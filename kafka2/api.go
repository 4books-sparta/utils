package kafka2

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

const (

	// Compression 	ratio 	CPU		Speed 		Network bandwidth usage
	// Gzip 	Highest Highest 	Slowest 	Lowest
	// Snappy 	Medium 	Moderate 	Moderate 	Medium
	// Lz4 		Low 	Lowest 		Fastest 	Highest
	// Zstd 	Medium 	Moderate 	Moderate 	Medium

	DEFAULT_COMPRESSION = "snappy"
)

type KafkaRecord kgo.Record

type kafkaConfig struct {
	clientID     string
	seeds        []string
	group        string
	topic        string
	verbose      bool
	saslEnabled  bool
	saslMech     string
	saslUser     string
	saslPassword string
	autocommit   bool
	compression  string
	atStart      bool
	atEnd        bool
	atTimestamp  int32
	balancer     string
	syncProducer bool
	partitioner  string
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
		balancer:     "sticky",
		syncProducer: false,
		partitioner:  "",
	}
}

type KafkaOption func(*kafkaConfig)

func Verbose(verbose bool) KafkaOption {
	return func(cfg *kafkaConfig) {
		cfg.verbose = verbose
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
