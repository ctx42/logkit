module example

go 1.24.0

require (
	github.com/ctx42/logkit v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/ctx42/testing v0.38.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/ctx42/logkit => ../..
