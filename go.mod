module github.com/rackspace-spot/spotctl

go 1.23.5

require (
	github.com/fatih/color v1.18.0
	github.com/rackspace-spot/spot-go-sdk v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.9.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.25.0 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
)

replace github.com/rackspace-spot/spot-go-sdk => /home/rajendra-gosavi/platform9/spot-go-sdk
