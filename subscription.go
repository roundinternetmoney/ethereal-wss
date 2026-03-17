package etherealWss

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

type EventData interface{}

type SymbolEvent struct {
	EventData
	T string `json:"type"`
	S string `json:"symbol"`
}
type SubaccountEvent struct {
	EventData
	T string `json:"type"`
	S string `json:"subaccountId"`
}

type Subscription[Proto proto.Message, Data EventData] struct {
	*websocket.Conn
	eventName string
	data      Proto
	Callback  func(Proto)

	eventType EventData
}

func NewSubscription[P proto.Message, D EventData](callback func(P)) *Subscription[P, D] {
	return &Subscription[P, D]{}
}

type Intent string

const (
	Sub   Intent = "subscribe"
	Unsub Intent = "unsubscribe"
)

type SubscriptionIntent[T EventData] struct {
	I Intent    `json:"event"`
	D EventData `json:"data"`
}

func (c *Subscription[_, EventData]) Subscribe(ctx context.Context) error {
	if bytes, err := json.Marshal(&SubscriptionIntent[EventData]{I: Sub, D: c.eventType}); err != nil {
		return err
	} else {
		return c.Write(ctx, websocket.MessageBinary, bytes)
	}
}

func (c *Subscription[_, EventData]) Unsubscribe(ctx context.Context) error {
	if bytes, err := json.Marshal(&SubscriptionIntent[EventData]{I: Unsub, D: c.eventType}); err != nil {
		return err
	} else {
		return c.Write(ctx, websocket.MessageBinary, bytes)
	}
}
