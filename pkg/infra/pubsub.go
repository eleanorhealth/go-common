package infra

import (
	"context"

	"cloud.google.com/go/pubsub"
)

type PubsubMessagePublisher interface {
	Publish(ctx context.Context, msg *pubsub.Message) error
}

type NopPublisher struct {
}

func (n *NopPublisher) Publish(ctx context.Context, msg *pubsub.Message) error {
	return nil
}

type PubsubPublisher struct {
	t *pubsub.Topic
}

func NewPubsubPublisher(t *pubsub.Topic) *PubsubPublisher {
	return &PubsubPublisher{
		t: t,
	}
}

func (p *PubsubPublisher) Publish(ctx context.Context, msg *pubsub.Message) error {
	res := p.t.Publish(ctx, msg)
	_, err := res.Get(ctx)

	return err
}

type PubsubMessageReceiver interface {
	Receive(ctx context.Context, f func(context.Context, *PubsubMessage)) error
}

type PubsubMessage struct {
	ID     string
	AckFn  func()
	NackFn func()
	Data   []byte
}

func (p *PubsubMessage) Ack() {
	p.AckFn()
}

func (p *PubsubMessage) Nack() {
	p.NackFn()
}

type NopReceiver struct {
}

func (n *NopReceiver) Receive(ctx context.Context, f func(context.Context, *PubsubMessage)) error {
	return nil
}

type PubsubReceiver struct {
	s *pubsub.Subscription
}

func NewPubsubReceiver(s *pubsub.Subscription) *PubsubReceiver {
	return &PubsubReceiver{
		s: s,
	}
}

func (p *PubsubReceiver) Receive(ctx context.Context, f func(context.Context, *PubsubMessage)) error {
	return p.s.Receive(ctx, func(pCtx context.Context, pMsg *pubsub.Message) {
		msg := &PubsubMessage{
			ID:     pMsg.ID,
			Data:   pMsg.Data,
			AckFn:  pMsg.Ack,
			NackFn: pMsg.Nack,
		}

		f(ctx, msg)
	})
}
