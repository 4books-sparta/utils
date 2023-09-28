package gc_pubsub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

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
	printError   bool
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

func (c *Client) PrintErrors() {
	c.printError = true
}

func (c *Client) HideErrors() {
	c.printError = false
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

	if c.activeTopics == nil {
		c.activeTopics = make(map[string]*pubsub.Topic)
	}

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
			c.Log(InfoLevel, "INFO: Iterator DONE")
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

func (c *Client) Publish(ctx context.Context, t *pubsub.Topic, msg string, key string) error {
	message := pubsub.Message{
		Data: []byte(msg),
	}
	if key != "" {
		message.OrderingKey = key
		t.EnableMessageOrdering = true
	} else {
		t.EnableMessageOrdering = true
	}
	c.SetActiveTopic(t)

	result := t.Publish(ctx, &message)

	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		if key != "" {
			t.ResumePublish(key) // Resume publishing for that key
		}
		c.ErrorLog("Publishing "+msg, err)
		return err
	}
	c.Log(InfoLevel, fmt.Sprintf("Published a message; msg ID: %v", id))

	return nil
}

func (c *Client) PublishToTopicID(ctx context.Context, topic, msg, key string) error {
	t := c.GetActiveTopic(topic)
	if key != "" {
		t.EnableMessageOrdering = true
	}
	return c.Publish(ctx, t, msg, key)
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

func (c *Client) PanicHandler(todo func(args any)) {
	r := recover()
	if r == nil {
		return // no panic underway
	}

	todo(r)

	// print debug stack
	debug.PrintStack()

	os.Exit(1)
}

func (c *Client) Consume(ctx context.Context, sub *pubsub.Subscription, handler func(...interface{}) error, handlerArgs []interface{}) error {
	return sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		if handleErr := handler(message.Data, handlerArgs); handleErr != nil {
			c.Log(ErrorLevel, "NACK "+handleErr.Error()+string(message.Data))
			fmt.Println("NACK ", handleErr.Error(), string(message.Data))
			message.Nack()
			return
		}
		message.Ack()
	})

}

func (c *Client) MustGetSub(ctx context.Context, subName string, topicName string, cancel func()) *pubsub.Subscription {
	sub := c.getSub(ctx, subName, cancel)
	if sub == nil {
		topic := c.GetTopic(topicName)
		err := c.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
			Topic:                         topic,
			AckDeadline:                   10 * time.Second,
			RetainAckedMessages:           true,
			RetentionDuration:             24 * 7 * time.Hour, // 1 week
			EnableMessageOrdering:         true,
			DeadLetterPolicy:              nil,
			TopicMessageRetentionDuration: 0,
			EnableExactlyOnceDelivery:     false,
		})
		if err != nil {
			cancel()
			panic(err)
		}
		//At this time subscription must exist
		sub = c.getSub(ctx, subName, cancel)
		if sub == nil {
			cancel()
			panic("unable-to-create-subscription")
		}
	}
	return sub
}

func (c *Client) getSub(ctx context.Context, subName string, cancel func()) *pubsub.Subscription {
	sub := c.Subscribe(subName)
	if sub == nil {
		cancel()
		c.Log(InfoLevel, "Creating "+subName)
		panic("no-subscription")
	}

	//Check if the sub exists on the server
	exists, err := sub.Exists(ctx)
	if err != nil {
		cancel()
		panic(err)
	}

	if !exists {
		return nil
	}

	return sub
}

func (c *Client) ExitHandler(ctx context.Context, cancel func()) {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
		fmt.Println("Graceful consumer shutdown")
		os.Exit(1)
	}()
}
