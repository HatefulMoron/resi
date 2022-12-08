module github.com/hatefulmoron/resi

go 1.16

require (
	github.com/golang/protobuf v1.5.2
	github.com/hatefulmoron/resi/protos v0.0.0-00010101000000-000000000000
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/nknorg/ncp-go v1.0.3
	github.com/nknorg/nkn-sdk-go v1.4.1
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.2
	google.golang.org/grpc v1.47.0
)

replace github.com/hatefulmoron/resi/protos => ./protos
