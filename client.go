package etherealWss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"roundinternet.money/pb-dex"
)

type Environment string

const (
	Testnet Environment = "wss://ws2.etherealtest.net/v1/stream"
	Mainnet Environment = "wss://ws2.ethereal.trade/v1/stream"
)

type Client struct {
	Con           *websocket.Conn
	conMu         *sync.Mutex
	env           Environment
	subscriptions []pb.EventType
	callbacks     map[string]func(proto.Message)
	hbCancel      context.CancelCauseFunc
	pbOpts        *protojson.UnmarshalOptions
}

func NewClient(parent context.Context, env Environment) *Client {
	ctx, cancel := context.WithCancelCause(parent)
	c, _, err := websocket.Dial(ctx, string(env), nil)
	if err != nil {
		log.Fatal(err)
	}

	cl := &Client{
		Con:           c,
		conMu:         &sync.Mutex{},
		env:           env,
		subscriptions: make([]pb.EventType, 0),
		callbacks:     make(map[string]func(proto.Message)),
		pbOpts:        &protojson.UnmarshalOptions{DiscardUnknown: true},
	}

	cl.keepalive(ctx, cancel)
	cl.hbCancel = cancel

	return cl
}

func (c *Client) Subscribe(ctx context.Context, event pb.EventType, to string) (err error) {
	var bytes []byte
	fmt.Println(to)
	if bytes, err = event.MarshalIntent(to, pb.Sub); err != nil {
		return
	}
	fmt.Println(string(bytes))
	if err = c.Req(ctx, bytes); err != nil {
		c.subscriptions = append(c.subscriptions, event)
	}
	return
}

func (c *Client) Unsubscribe(ctx context.Context, event pb.EventType, to string) (err error) {
	var bytes []byte
	if bytes, err = event.MarshalIntent(to, pb.Unsub); err != nil {
		return err
	}
	return c.Req(ctx, bytes)
}

func (c *Client) SubscribeWithCallback(ctx context.Context, event pb.EventType, to string, cb func(proto.Message)) (err error) {
	if err = c.Subscribe(ctx, event, to); err == nil {
		c.callbacks[event.EventName()] = cb
	}
	return
}

func (c *Client) OnEvent(event pb.EventType, cb func(proto.Message)) {
	c.callbacks[event.EventName()] = cb
}

func (c *Client) Req(ctx context.Context, payload []byte) (err error) {
	return c.Con.Write(ctx, websocket.MessageBinary, payload)
}

type wssMsg struct {
	Event string          `json:"e"`
	Ts    int64           `json:"t"`
	Data  json.RawMessage `json:"data"`
}

func (c *Client) Listen(parent context.Context) error {
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	defer c.Close()

	for {
		_, data, err := c.Con.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return context.Cause(ctx)
			}
			cancel(err)
			return err
		}

		var e pb.EventMessage
		if err := c.pbOpts.Unmarshal(data, &e); err != nil {
			if status := new(pb.WebsocketStatus); c.pbOpts.Unmarshal(data, status) == nil {
				if !status.Ok {
					fmt.Println(status.Code)
				}
			} else {
				panic(err)
			}
			continue
		}

		if cb, ok := c.callbacks[e.E]; ok {
			event := pb.EventEnum(e.E)
			if err = event.UnmarshalToCallback(data, cb); err != nil {
				cancel(err)
				return err
			}
		}
	}
}

func (c *Client) keepalive(ctx context.Context, cancel context.CancelCauseFunc) {
	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				// Ping will return error if connection is dead
				if err := c.Con.Ping(ctx); err != nil {
					cancel(err)
					return
				}
			}
		}
	}()
}

func (c *Client) Resubscribe(parent context.Context) error {
	c.Close()

	c.conMu.Lock()
	defer c.conMu.Unlock()

	ctx, cancel := context.WithCancelCause(parent)

	// replace con and restart listener with new context
	var err error
	c.Con, _, err = websocket.Dial(ctx, string(c.env), nil)
	if err != nil {
		cancel(err)
		return err
	}
	c.hbCancel = cancel
	c.keepalive(ctx, cancel)

	return nil
}

// func (c *Client) UnsubscribeAll(ctx context.Context) (err error) {
// 	for _, s := range c.subscriptions {
// 		if err = c.Unsubscribe(ctx, s); err != nil {
// 			return err
// 		}
// 	}
// 	return
// }

func (c *Client) Close() {
	c.conMu.Lock()
	defer c.conMu.Unlock()
	if c.hbCancel != nil {
		c.hbCancel(nil)
	}
	if c.Con != nil {
		c.Con.Close(websocket.StatusNormalClosure, "closing")
	}
}

/*

func (c *Client) SubscribeBook(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SymbolEvent{
		T: "L2Book",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeBook(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SymbolEvent{
		T: "L2Book",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeMarketPrice(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SymbolEvent{
		T: "MarketPrice",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeMarketPrice(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SymbolEvent{
		T: "MarketPrice",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeFill(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SymbolEvent{
		T: "TradeFill",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeFill(ctx context.Context, symbol string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SymbolEvent{
		T: "TradeFill",
		S: symbol,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeLiquidation(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SubaccountEvent{
		T: "SubaccountLiquidation",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeLiquidation(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SubaccountEvent{
		T: "SubaccountLiquidation",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeOrderUpdate(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SubaccountEvent{
		T: "OrderUpdate",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeOrderUpdate(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SubaccountEvent{
		T: "OrderUpdate",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeOrderFill(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SubaccountEvent{
		T: "OrderFill",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeOrderFill(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SubaccountEvent{
		T: "OrderFill",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) SubscribeTokenTransfer(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalSubscribe(&SubaccountEvent{
		T: "TokenTransfer",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) UnsubscribeTokenTransfer(ctx context.Context, subaccountUuid string) (err error) {
	var b []byte
	if b, err = marshalUnsubscribe(&SubaccountEvent{
		T: "TokenTransfer",
		S: subaccountUuid,
	}); err != nil {
		return err
	}
	return c.req(ctx, b)
}

func (c *Client) OnBook(callback func(*pb.L2Book)) {
	c.bookHandler = callback
}

func (c *Client) OnPrice(callback func(*pb.MarketPrice)) {
	c.priceHandler = callback
}

func (c *Client) OnTradeFill(callback func(*pb.TradeFillEvent)) {
	c.tradeFillHandler = callback
}

func (c *Client) OnLiquidation(callback func(*pb.SubaccountLiquidationEvent)) {
	c.liquidationHandler = callback
}

func (c *Client) OnOrderUpdate(callback func(*pb.OrderUpdateEvent)) {
	c.orderUpdateHandler = callback
}

func (c *Client) OnOrderFill(callback func(*pb.OrderFillEvent)) {
	c.orderFillHandler = callback
}

func (c *Client) OnTransfer(callback func(*pb.Transfer)) {
	c.transferHandler = callback
}

*/
