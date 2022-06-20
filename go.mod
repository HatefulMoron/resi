module resi

go 1.16

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hatefulmoron/resi/protos v0.0.0-00010101000000-000000000000
	github.com/nknorg/nkn-sdk-go v1.4.1
	github.com/stretchr/testify v1.7.2
	github.com/urfave/cli/v2 v2.10.2 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	google.golang.org/grpc v1.47.0
)

replace github.com/hatefulmoron/resi/protos => ./protos
