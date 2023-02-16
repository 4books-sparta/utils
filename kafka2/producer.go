package kafka2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/4books-sparta/utils"
)

type KafkaProducer struct {
	client  *kgo.Client
	cfg     *kafkaConfig
	opts    []kgo.Opt
	Ch      chan *kgo.Record
	wg      sync.WaitGroup
	Verbose bool
}

func die(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
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

	if k.cfg.verbose {
		utils.PrintVarDump("CONFIG ", k.cfg)
	}

	if nop, err := KafkaAuth(k.cfg); err == nil && nop != nil {
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

func (k *KafkaProducer) Ping() {
	fmt.Println("PING")
	err := k.client.Ping(context.Background())
	if err != nil {
		fmt.Println("Error pinging", err)
	}
	fmt.Println("Pong")
}

func (k *KafkaProducer) Start(cb func(r *kgo.Record, err error)) error {
	log.Printf("Starting kafka producer for topic '%s' to brokers %+v. Syncronous: %v Partitioner: %s Compression: %s SASL: %s ",
		k.cfg.topic,
		k.cfg.seeds,
		k.cfg.syncProducer,
		k.cfg.partitioner,
		k.cfg.compression,
		k.cfg.saslMech,
	)
	var err error

	if k.cfg.verbose {
		fmt.Println("Using TLS? ", k.cfg.dialTLS != nil)
	}
	k.client, err = kgo.NewClient(k.opts...)
	if err != nil {
		log.Printf("error initializing Kafka Producer: %v\n", err)
		return err
	}

	if !k.cfg.syncProducer {
		k.wg.Add(1)

		go func() {
			defer k.wg.Done()
			defer k.Stop()
			for {
				select {
				case msg, ok := <-k.Ch:
					if !ok {
						log.Printf("Shutting down Kafka Producer\n")
						return
					}
					k.client.Produce(context.Background(), msg, cb)
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
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
	if !k.cfg.syncProducer {
		if k.Verbose {
			fmt.Println("Producing async")
		}
		k.Ch <- rec
		return nil
	} else {
		if k.Verbose {
			fmt.Println("Producing sync")
		}
		res := k.client.ProduceSync(context.Background(), rec)
		if err := res.FirstErr(); err != nil {
			log.Printf("Error producing")
			return err
		}
		if k.Verbose {
			fmt.Println("Produced")
			utils.PrintVarDump("RES", res)
		}
		return nil
	}
}
