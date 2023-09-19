package gc_pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

const (
	ErrorLevel = "ERROR"
	InfoLevel  = "INFO"
)

type Client struct {
	client *pubsub.Client
	debug  bool
}

type ClientConfig struct {
	ProjectId string
}

func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	cl, err := pubsub.NewClient(ctx, cfg.ProjectId)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: cl,
		debug:  false,
	}, nil
}

func (c *Client) StartDebug() {
	c.debug = true
}

func (c *Client) StopDebug() {
	c.debug = false
}

func (c *Client) CreateTopicIfNotExists(ctx context.Context, topic string) (*pubsub.Topic, error) {
	t := c.client.Topic(topic)
	ok, err := t.Exists(ctx)
	if err != nil {
		c.ErrorLog("Failed to check existence of the topic", err)
		return nil, err
	}
	if ok {
		return t, nil
	}

	return c.CreateTopic(ctx, topic)
}

func (c *Client) CreateTopic(ctx context.Context, topic string) (*pubsub.Topic, error) {
	t, err := c.client.CreateTopic(ctx, topic)
	if err != nil {
		c.ErrorLog("Failed to create the topic", err)
		return nil, err
	}
	c.Log(InfoLevel, fmt.Sprintf("Topic created: %v", t))
	return t, nil
}

func (c *Client) CreateSubscription(ctx context.Context, name string, topic *pubsub.Topic) error {
	sub, err := c.client.CreateSubscription(ctx, name, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		c.ErrorLog(fmt.Sprintf("Error creating subscription: %v\n", sub), err)
		return err
	}
	c.Log(InfoLevel, fmt.Sprintf("Created subscription: %v", sub))
	return nil
}

func (c *Client) HandleMessages(ctx context.Context, name string, handle func(context.Context, *pubsub.Message) error) error {
	var mu sync.Mutex
	sub := c.client.Subscription(name)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		failErr := handle(ctx, msg)
		if failErr != nil {
			c.ErrorLog("cant-handle-message"+msg.ID, failErr)
			msg.Nack()
		} else {
			msg.Ack()
		}

		c.Log(InfoLevel, fmt.Sprintf("Got message => %q\n", string(msg.Data)))
		defer mu.Unlock()
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListTopics(ctx context.Context) ([]*pubsub.Topic, error) {
	var topics []*pubsub.Topic
	it := c.client.Topics(ctx)
	for {
		topic, err := it.Next()
		if errors.Is(err, iterator.Done) {
			if c.debug {
				fmt.Println("INFO: Iterator DONE")
			}
			break
		}
		if err != nil {
			c.ErrorLog("Iteration error", err)
			return nil, err
		}
		topics = append(topics, topic)
		c.Log(InfoLevel, " - topic: "+topic.String())

	}
	return topics, nil
}

func (c *Client) Publish(ctx context.Context, topic, msg string) error {
	t := c.client.Topic(topic)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		c.ErrorLog("Publishing "+msg, err)
		return err
	}
	c.Log(InfoLevel, fmt.Sprintf("Published a message; msg ID: %v", id))

	return nil
}

func (c *Client) Log(level, str string) {
	if !c.debug {
		return
	}

	fmt.Printf("%s: %s\n", level, str)
}

func (c *Client) ErrorLog(str string, err error) {
	if !c.debug {
		return
	}

	fmt.Printf("%s: %s =>  %v\n", ErrorLevel, str, err)
}
