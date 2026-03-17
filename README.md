# Golang Websocket Client for Ethereal API

[![Go Reference](https://pkg.go.dev/badge/roundinternet.money/ethereal-wss.svg)](https://pkg.go.dev/roundinternet.money/ethereal-wss)


## Features
- Protobuf support.
- Minimal dependencies

## Getting started

- Requires Go 1.25+.
- Install from GitHub: `go get github.com/Round-Internet-Money/ethereal-wss`

## Example Usage

From the client directory:

- `make listen_all`
- `bin/listen_to_everything`

## Modifying the package
- This client depends on protobuf wrappers from [github.com/Round-Internet-Money/pb-dex](https://github.com/Round-Internet-Money/pb-dex).
- If you want to extend the `.proto` files directly, see the Buf module at [buf.build/round-internet-money/dex](https://buf.build/round-internet-money/dex).
- Otherwise, use or fork [github.com/Round-Internet-Money/pb-dex](https://github.com/Round-Internet-Money/pb-dex).

Contributing
-------------
Contributions are welcome! Please open issues or pull requests as needed.


## Todo

- Add a `resubscribe` intent helper.
