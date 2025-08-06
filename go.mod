module github.com/rackerlabs/spot-cli

go 1.23.5

require (
	github.com/rackerlabs/spot-sdk/rxtspot v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.9.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
)

replace github.com/rackerlabs/spot-sdk/rxtspot => ../rxtspot
