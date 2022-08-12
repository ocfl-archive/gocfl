module go.ub.unibas.ch/gocfl/v2

go 1.19

replace (
	go.ub.unibas.ch/gocfl/v2 => ./
)

require (
	emperror.dev/emperror v0.33.0
	github.com/je4/utils/v2 v2.0.6
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
)

require (
	emperror.dev/errors v0.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20211102061401-a2f17f7b995c // indirect
)
