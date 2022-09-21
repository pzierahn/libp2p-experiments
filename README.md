# libp2p-experiments

```shell
rm -rf bin/
mkdir bin

GOOS=linux GOARCH=amd64 go build -o bin/relay cmd/relay/relay.go

scp bin/relay ubuntu@ec2-3-75-220-204.eu-central-1.compute.amazonaws.com:~/
```