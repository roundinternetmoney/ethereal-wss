module roundinternet.money/go-ethereal-websocket

go 1.25.0

require (
	github.com/coder/websocket v1.8.14
	github.com/joho/godotenv v1.5.1
	google.golang.org/protobuf v1.36.11
	roundinternet.money/ethereal-wss v0.0.0-00010101000000-000000000000
	roundinternet.money/pb-dex v0.0.0-20260317010626-cf790bf095f4
)

require buf.build/gen/go/round-internet-money/dex/protocolbuffers/go v1.36.11-20260317005403-e1bf6fd9924d.1 // indirect

replace roundinternet.money/ethereal-wss => ../ethereal-wss
