package kafka2

import (
	"context"
	"errors"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type KafkaProducer struct {
	client *kgo.Client
	cfg    *kafkaConfig
	opts   []kgo.Opt
	Ch     chan *kgo.Record
	wg     sync.WaitGroup
}

func KafkaProducerCreate(opts ...KafkaOption) (*KafkaProducer, error) {
	k := &KafkaProducer{
		Ch:  make(chan *kgo.Record),
		cfg: NewDefaultConfig(),
	}

	for _, opt := range opts {
		opt(k.cfg)
	}

	kopts := []kgo.Opt{
		kgo.ClientID(k.cfg.clientID),
		kgo.SeedBrokers(k.cfg.seeds...),
		kgo.DefaultProduceTopic(k.cfg.topic),
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
	case "snappy":
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

	switch strings.ToLower(k.cfg.partitioner) {
	case "round-robin", "":
		kopts = append(kopts, kgo.RecordPartitioner(kgo.RoundRobinPartitioner()))
	case "sticky":
		kopts = append(kopts, kgo.RecordPartitioner(kgo.StickyPartitioner()))
	default:
		e := errors.New("unrecognized partitioner " + k.cfg.partitioner)
		log.Printf("Error: %s", e.Error())
		return nil, e
	}

	if k.cfg.verbose {
		kopts = append(kopts,
			kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, nil)),
		)
	}

	// DisableAutoCommit should not be used.

	k.opts = kopts

	return k, nil
}

func (k *KafkaProducer) Start() error {
	log.Printf("Starting kafka producer for topic '%s' to brokers %+v. Syncronous: %v Partitioner: %s Compression: %s",
		k.cfg.topic,
		k.cfg.seeds,
		k.cfg.syncProducer,
		k.cfg.partitioner,
		k.cfg.compression,
	)
	var err error
	k.client, err = kgo.NewClient(k.opts...)
	if err != nil {
		log.Printf("error initializing Kafka Producer: %v\n", err)
		return err
	}

	if !k.cfg.syncProducer {
		k.wg.Add(1)

		go func() {
			defer k.wg.Done()
			for {
				select {
				case msg, ok := <-k.Ch:
					if !ok {
						log.Printf("Shutting down Kafka Producer\n")
						return
					}
					k.client.Produce(context.Background(), msg, func(r *kgo.Record, err error) {
						if err != nil {
							log.Printf("produce error: %s", err.Error())
							panic("Producing returned an error")
						}
					})
				}
			}
		}()
	}

	return nil
}

func (k *KafkaProducer) Stop() error {
	close(k.Ch)
	k.wg.Wait()
	k.client.Flush(context.Background())
	//k.client.Close()
	k.client.CloseAllowingRebalance()
	return nil
}

func (k *KafkaProducer) Send(key []byte, value []byte) error {
	rec := &kgo.Record{
		Topic:     k.cfg.topic,
		Key:       []byte(key),
		Value:     []byte(value),
		Timestamp: time.Now(),
	}
	if !k.cfg.syncProducer {
		k.Ch <- rec
		return nil
	} else {
		res := k.client.ProduceSync(context.Background(), rec)
		if err := res.FirstErr(); err != nil {
			log.Printf("Error producing")
			return err
		}
		return nil
	}
}
