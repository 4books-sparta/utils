package kafka2

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testKafkaConsumer(t *testing.T) {

	c, err := KafkaConsumerCreate(
		Verbose(false),
		ClientID("Suka"),
		Compression("none"),
		Autocommit(false),
		Seeds("172.25.17.18:9092"),
		Group("dbl2-test"),
		Topic("dbl2-test"),
		SASL("SCRAM-SHA-512", "dbladmin", "Ama4phah6xoeW4aelufeeg7a"),
		AtStart(),
		AtEnd(),
		AtTimestamp(-1),
	)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	//assert.Equal(t, tv.out, out.String())

	go func() {
		for {
			select {
			case msg := <-c.Ch:
				fmt.Printf("> %s\n", string(msg.Value))
			}
		}
	}()

	_ = c.Start()
	time.Sleep(5 * time.Second)
	_ = c.Commit()
	_ = c.Stop()
}
