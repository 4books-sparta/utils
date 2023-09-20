package gc_pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

const (
	ErrorLevel = "ERROR"
	InfoLevel  = "INFO"
)

type Client struct {
	client       *pubsub.Client
	debug        bool
	activeTopics map[string]*pubsub.Topic
	mu           sync.Mutex
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

func (c *Client) SetActiveTopic(t *pubsub.Topic) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.activeTopics[t.ID()]; !ok {
		c.activeTopics[t.ID()] = t
	}
}

func (c *Client) StopActiveTopics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.activeTopics == nil {
		return
	}

	for _, topic := range c.activeTopics {
		topic.Stop()
	}
}

func (c *Client) GetActiveTopics() map[string]*pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.activeTopics
}

func (c *Client) GetActiveTopic(topic string) *pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.activeTopics[topic]; !ok {
		c.activeTopics[topic] = c.client.Topic(topic)
	}

	return c.activeTopics[topic]
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

func (c *Client) StartDebug() {
	c.debug = true
}

func (c *Client) StopDebug() {
	c.debug = false
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

func (c *Client) CreateSubscription(ctx context.Context, name string, cfg pubsub.SubscriptionConfig) error {
	sub, err := c.client.CreateSubscription(ctx, name, cfg)
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
	c.Log(InfoLevel, "ListTopics: ")
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
		c.Log(InfoLevel, " Found - topic: "+topic.String())

	}
	return topics, nil
}

func (c *Client) Publish(ctx context.Context, t *pubsub.Topic, msg string) error {
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	c.SetActiveTopic(t)
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

func (c *Client) PublishToTopicID(ctx context.Context, topic, msg string) error {
	t := c.GetActiveTopic(topic)
	return c.Publish(ctx, t, msg)
}

func (c *Client) Subscribe(name string) *pubsub.Subscription {
	return c.client.Subscription(name)
}

func (c *Client) PubSubClient() *pubsub.Client {
	return c.client
}

func (c *Client) GetTopic(name string) *pubsub.Topic {
	return c.client.Topic(name)
}
