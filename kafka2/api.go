package kafka2

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

const (

	// Compression 	ratio 	CPU		Speed 		Network bandwidth usage
	// Gzip 	Highest Highest 	Slowest 	Lowest
	// Snappy 	Medium 	Moderate 	Moderate 	Medium
	// Lz4 		Low 	Lowest 		Fastest 	Highest
	// Zstd 	Medium 	Moderate 	Moderate 	Medium

	DEFAULT_COMPRESSION            = "snappy"
	SASL_MECHANISM_IAM             = "MSK_IAM_PLAIN"
	SASL_MECHANISM_PLAIN           = "PLAIN"
	SASL_MECHANISM_SHA_512         = "SCRAM-SHA-512"
	SASL_MECHANISM_SHA_256         = "SCRAM-SHA-256"
	PARTITIONER_STICKY             = "sticky"
	PARTITIONER_COOPERATIVE_STICKY = "cooperative-sticky"
	PARTITIONER_ROUND_ROBIN        = "round-robin"
)

type KafkaRecord kgo.Record

type kafkaConfig struct {
	clientID         string
	seeds            []string
	group            string
	topic            string
	verbose          bool
	saslEnabled      bool
	saslMech         string
	saslUser         string
	saslPassword     string
	autocommit       bool
	commitOnlyMarked bool
	compression      string
	atStart          bool
	atEnd            bool
	atTimestamp      int32
	balancer         string
	syncProducer     bool
	partitioner      string
	dialTLS          *tls.Config
	onRevoked        func(context.Context, *kgo.Client, map[string][]int32)
}

func NewDefaultConfig() *kafkaConfig {
	return &kafkaConfig{
		clientID:     "",
		seeds:        []string{},
		group:        "",
		topic:        "",
		verbose:      false,
		saslEnabled:  false,
		saslMech:     "",
		saslUser:     "",
		saslPassword: "",
		autocommit:   true,
		compression:  DEFAULT_COMPRESSION,
		atStart:      false,
		atEnd:        true,
		atTimestamp:  0,
		balancer:     PARTITIONER_STICKY,
		syncProducer: false,
		partitioner:  PARTITIONER_STICKY,
	}
}

type KafkaOption func(*kafkaConfig)

func Verbose(verbose bool) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.verbose = verbose
	}
}

func OnRevoked(onRev func(context.Context, *kgo.Client, map[string][]int32)) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.onRevoked = onRev
	}
}

func Autocommit(autocommit bool) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.autocommit = autocommit
	}
}

func SyncProducer(val bool) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.syncProducer = val
	}
}

func AutocommitOnlyMarked(val bool) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.commitOnlyMarked = val
	}
}

func ClientID(client string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.clientID = client
	}
}

func Seeds(seeds ...string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.seeds = append(cfg.seeds[:0], seeds...)
	}
}

func Group(group string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.group = group
	}
}

func Balancer(bname string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.balancer = bname
	}
}

func Topic(topic string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.topic = topic
	}
}

func SASL(mech, user, pass string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.saslMech = mech
		cfg.saslUser = user
		cfg.saslPassword = pass
		cfg.saslEnabled = true
	}
}

func Compression(comp string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.compression = comp
	}
}

func AtStart() KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.atStart = true
		cfg.atEnd = false
		cfg.atTimestamp = -1
	}
}

func AtEnd() KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.atStart = false
		cfg.atEnd = true
		cfg.atTimestamp = -1
	}
}

func UseTLS(serverName string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.dialTLS = &tls.Config{
			ServerName: serverName,
		}
	}
}

func AtTimestamp(val int32) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.atStart = false
		cfg.atEnd = false
		cfg.atTimestamp = val
	}
}

func Partitioner(val string) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.partitioner = val
	}
}

func KafkaAuth(cfg *kafkaConfig) (kgo.Opt, error) {
	if cfg.saslEnabled {
		if cfg.verbose {
			fmt.Println("Using", cfg.saslMech)
		}
		switch cfg.saslMech {
		case SASL_MECHANISM_SHA_256:
			return kgo.SASL(scram.Auth{
				User: cfg.saslUser,
				Pass: cfg.saslPassword,
			}.AsSha256Mechanism()), nil
		case SASL_MECHANISM_SHA_512:
			return kgo.SASL(scram.Auth{
				User: cfg.saslUser,
				Pass: cfg.saslPassword,
			}.AsSha512Mechanism()), nil
		case SASL_MECHANISM_PLAIN:
			return kgo.SASL(plain.Auth{
				User: cfg.saslUser,
				Pass: cfg.saslPassword,
			}.AsMechanism()), nil
		case SASL_MECHANISM_IAM:
			sess, err := session.NewSession()
			if err != nil {
				die("unable to initialize aws session: %v", err)
			}
			return kgo.SASL(aws.ManagedStreamingIAM(func(ctx context.Context) (aws.Auth, error) {
				val, err := sess.Config.Credentials.GetWithContext(ctx)
				if err != nil {
					return aws.Auth{}, err
				}
				if cfg.verbose {
					fmt.Println("AK => ", val.AccessKeyID)
				}
				return aws.Auth{
					AccessKey:    val.AccessKeyID,
					SecretKey:    val.SecretAccessKey,
					SessionToken: val.SessionToken,
					UserAgent:    "franz-go/creds_test/v1.0.0",
				}, nil
			})), nil
		default:
			return nil, errors.New("SASL mechanism not supported")
		}
	}
	return nil, nil
}
