package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	ws "roundinternet.money/ethereal-wss"
	pb "roundinternet.money/pb-dex"

	"google.golang.org/protobuf/proto"
)

const symbol = "BTCUSD"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := ws.NewClient(ctx, ws.Testnet)
	defer client.Close()

	events := []pb.EventType{
		pb.EventType_EVENT_TYPE_L2_BOOK,
		pb.EventType_EVENT_TYPE_TICKER,
		pb.EventType_EVENT_TYPE_TRADE_FILL,
	}

	for _, event := range events {
		if err := client.SubscribeWithCallback(ctx, event, symbol, func(msg proto.Message) {
			fmt.Printf("[%s] %T %v\n", event.EventName(), msg, msg)
		}); err != nil {
			log.Fatalf("subscribe %s: %v", event.EventName(), err)
		}
	}

	payload, err := pb.MarshalIntentForMessage(&pb.Ticker{}, symbol, pb.Sub)
	if err != nil {
		log.Fatalf("build ticker subscription: %v", err)
	}
	fmt.Printf("ticker subscribe payload: %s\n", payload)

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Listen(ctx) // blocking
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errCh:
		if err != nil && ctx.Err() == nil {
			log.Fatal(err)
		}
	}
}
