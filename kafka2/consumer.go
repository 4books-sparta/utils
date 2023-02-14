package kafka2

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

const (
	CompressionSnappy = "snappy"
)

type KafkaConsumer struct {
	client             *kgo.Client
	cfg                *kafkaConfig
	opts               []kgo.Opt
	Ch                 chan *KafkaRecord
	offsets            map[int32]int64
	current            map[string]map[int32]kgo.EpochOffset
	commitLock         sync.Mutex
	lastCommit         *time.Time
	uncommittedRecords map[int32]kgo.EpochOffset
}

func KafkaConsumerCreate(opts ...KafkaOption) (*KafkaConsumer, error) {
	k := &KafkaConsumer{
		Ch:                 make(chan *KafkaRecord),
		offsets:            make(map[int32]int64),
		cfg:                NewDefaultConfig(),
		uncommittedRecords: make(map[int32]kgo.EpochOffset, 0),
	}

	for _, opt := range opts {
		opt(k.cfg)
	}

	kopts := []kgo.Opt{
		kgo.ClientID(k.cfg.clientID),
		kgo.SeedBrokers(k.cfg.seeds...),
		kgo.ConsumerGroup(k.cfg.group),
		kgo.ConsumeTopics(k.cfg.topic),
		kgo.FetchIsolationLevel(kgo.ReadCommitted()), // only read messages that have been written as part of committed transactions
	}

	if k.cfg.saslEnabled {
		var nop kgo.Opt

		switch k.cfg.saslMech {
		case "SCRAM-SHA-256":
			nop = kgo.SASL(scram.Auth{
				User: k.cfg.saslUser,
				Pass: k.cfg.saslPassword,
			}.AsSha256Mechanism())
		case "SCRAM-SHA-512":
			nop = kgo.SASL(scram.Auth{
				User: k.cfg.saslUser,
				Pass: k.cfg.saslPassword,
			}.AsSha512Mechanism())
		case "PLAIN":
			nop = kgo.SASL(plain.Auth{
				User: k.cfg.saslUser,
				Pass: k.cfg.saslPassword,
			}.AsMechanism())
		default:
			return nil, errors.New("SASL Mech not supported")
		}
		kopts = append(kopts, nop)
	}

	switch strings.ToLower(k.cfg.compression) {
	case "", "none":
		kopts = append(kopts, kgo.ProducerBatchCompression(kgo.NoCompression()))
	case "gzip":
		kopts = append(kopts, kgo.ProducerBatchCompression(kgo.GzipCompression()))
	case CompressionSnappy:
		kopts = append(kopts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	case "lz4":
		kopts = append(kopts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
	case "zstd":
		kopts = append(kopts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
	default:
		e := errors.New("unrecognized compression " + k.cfg.compression)
		log.Printf("Error: %s", e.Error())
		return nil, e
	}

	if !k.cfg.autocommit {
		kopts = append(kopts,
			kgo.DisableAutoCommit(),
		)
	}

	var off kgo.Offset
	if k.cfg.atStart {
		off = kgo.NewOffset().AtStart()
	} else if k.cfg.atEnd {
		off = kgo.NewOffset().AtEnd()
	} else if k.cfg.atTimestamp > 0 {
		off = kgo.NewOffset().WithEpoch(k.cfg.atTimestamp)
	} else {
		off = kgo.NoResetOffset()
		//off = kgo.NewOffset().AfterMilli(1675599465000)
	}
	kopts = append(kopts,
		kgo.ConsumeResetOffset(off),
	)

	if k.cfg.verbose {
		kopts = append(kopts,
			kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, nil)),
		)
	}

	var balancer kgo.GroupBalancer

	switch k.cfg.balancer {
	case "range":
		balancer = kgo.RangeBalancer()
	case "roundrobin":
		balancer = kgo.RoundRobinBalancer()
	case "sticky":
		balancer = kgo.StickyBalancer()
	case "cooperative-sticky":
		balancer = kgo.CooperativeStickyBalancer()
	default:
		log.Fatalf("unrecognized group balancer: %s", k.cfg.balancer)
	}

	kopts = append(kopts,
		kgo.Balancers(balancer),
	)

	k.opts = kopts

	return k, nil
}

func (k *KafkaConsumer) Start() error {
	log.Printf("Starting kafka consumer for topic '%s' to brokers %+v. Autocommit: %v",
		k.cfg.topic,
		k.cfg.seeds,
		k.cfg.autocommit,
	)
	var err error
	k.client, err = kgo.NewClient(k.opts...)
	if err != nil {
		log.Printf("error initializing Kafka Consumer: %v\n", err)
		return err
	}

	k.current = k.client.MarkedOffsets()

	go k.consume()

	return nil
}

func (k *KafkaConsumer) Stop() error {
	k.client.CloseAllowingRebalance()
	return nil
}

func (k *KafkaConsumer) MarkOffset(row *KafkaRecord) {
	k.commitLock.Lock()
	defer k.commitLock.Unlock()

	k.uncommittedRecords[row.Partition] = kgo.EpochOffset{
		Offset: row.Offset + 1,
	}
}

func (k *KafkaConsumer) Rollback() {
	k.client.CommitOffsetsSync(context.Background(), k.current, nil)
}

func (k *KafkaConsumer) Commit() error {
	if k.cfg.autocommit {
		return nil
	}

	k.commitLock.Lock()
	defer k.commitLock.Unlock()

	if k.uncommittedRecords == nil || len(k.uncommittedRecords) == 0 {
		//Nothing to be committed
		return nil
	}

	/*po := make(map[int32]kgo.EpochOffset)
	for p, o := range k.offsets {
		po[p] = kgo.EpochOffset{
			Offset: o + 1,
		}
	}*/

	now := time.Now()
	uncommitted := make(map[string]map[int32]kgo.EpochOffset)
	uncommitted[k.cfg.topic] = k.uncommittedRecords
	k.client.CommitOffsetsSync(context.Background(), uncommitted, func(cc *kgo.Client, oo *kmsg.OffsetCommitRequest, rr *kmsg.OffsetCommitResponse, err error) {
		if err != nil {
			log.Printf("Error committing offsets: %s", err.Error())
		}
	})
	k.current = k.client.MarkedOffsets()

	//Reset commits
	k.uncommittedRecords = make(map[int32]kgo.EpochOffset, 0)
	k.lastCommit = &now

	return nil
}

func (k *KafkaConsumer) CommitAfter(d time.Duration) error {
	now := time.Now()
	if k.lastCommit != nil && k.lastCommit.Add(d).After(now) {
		//Too early
		return nil
	}

	defer func() {
		k.lastCommit = &now
	}()

	return k.Commit()
}

func (k *KafkaConsumer) consume() {
	for {
		fetches := k.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			// TODO Close the chan
			return
		}

		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("Fetch error topic %s partition %d: %v", t, p, err)
			if strings.Contains(err.Error(), "the client consumed to offset") && strings.Contains(err.Error(), "but was reset to offset") {
				return
			}
			os.Exit(1)
		})

		fetches.EachRecord(func(r *kgo.Record) {
			kr := &KafkaRecord{
				Key:       r.Key,
				Value:     r.Value,
				Topic:     r.Topic,
				Partition: r.Partition,
				Offset:    r.Offset,
				Timestamp: r.Timestamp,
			}
			k.Ch <- kr
			k.commitLock.Lock()
			k.offsets[r.Partition] = r.Offset
			k.commitLock.Unlock()
		})
	}
}
