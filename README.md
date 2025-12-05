# cs249-dht
Final project for CS249 class.

## Requirements
1. Install golang version 1.24.3
2. Install required external dependencies: `go get -d ./...`

## Running the code
From a terminal, start the program using:
`go run .`

View command line options by running `go run . -help`

To begin, run a bootstrap node in one terminal or machine: `go run . -b`
This will start a node on localhost and port 8090.

To add another node, run: `go run . -p=8091` in another terminal. It should automatically bootstrap to the bootstrap address.

If the bootstrap address is remote, e.g. 1.2.3.4:8090, run `go run . -p=<local port> -ba="1.2.3.4" -bp=8090`
