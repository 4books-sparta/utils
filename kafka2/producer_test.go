package kafka2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testKafkaProducer(t *testing.T) {

	c, err := KafkaProducerCreate(
		Verbose(false),
		ClientID("Suka"),
		Compression("none"),
		Autocommit(false),
		Seeds("172.25.17.18:9092"),
		Group("dbl2-test"),
		Topic("dbl2-test"),
		SASL("SCRAM-SHA-512", "dbladmin", "Ama4phah6xoeW4aelufeeg7a"),
		AtStart(),
		SyncProducer(true),
		Partitioner(PARTITIONER_STICKY),
	)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	//assert.Equal(t, tv.out, out.String())

	c.Start(nil)

	for t := 0; t < 10; t++ {
		fmt.Printf("<< %d\n", t)
		if err := c.Send(nil, []byte(fmt.Sprintf("%d", t))); err != nil {
			fmt.Printf("Error sending data to producer: %s\n", err.Error())
		}
	}

	c.Stop()
}
