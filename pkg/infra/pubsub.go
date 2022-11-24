package infra

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/eleanorhealth/go-common/pkg/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PubSubMessagePublisher interface {
	Publish(ctx context.Context, msg *pubsub.Message) error
}

type NopPublisher struct {
}

func (n *NopPublisher) Publish(ctx context.Context, msg *pubsub.Message) error {
	return nil
}

type PubSubPublisher struct {
	t *pubsub.Topic
}

func NewPubSubPublisher(t *pubsub.Topic) *PubSubPublisher {
	return &PubSubPublisher{
		t: t,
	}
}

func (p *PubSubPublisher) Publish(ctx context.Context, msg *pubsub.Message) error {
	res := p.t.Publish(ctx, msg)
	_, err := res.Get(ctx)

	return err
}

type PubSubMessage struct {
	ID         string
	Data       []byte
	Attributes map[string]string
}

const (
	pubsubLocalAckDeadline      = 10 * time.Second
	pubsubLocalExpirationPolicy = 25 * time.Hour
)

func InitLocalPubsub(ctx context.Context, client *pubsub.Client, topicID, subID string) (*pubsub.Topic, *pubsub.Subscription, error) {
	_, err := client.CreateTopic(ctx, topicID)
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return nil, nil, errs.Wrap(err, "creating topic")
	}

	topic := client.Topic(topicID)

	_, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
		Topic:            topic,
		AckDeadline:      pubsubLocalAckDeadline,
		ExpirationPolicy: pubsubLocalExpirationPolicy,
	})
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return nil, nil, errs.Wrap(err, "creating subscription")
	}

	sub := client.Subscription(subID)

	return topic, sub, nil
}
