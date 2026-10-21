module example

go 1.26.0

require (
	github.com/ctx42/logkit v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.35.1
	github.com/sirupsen/logrus v1.10.2
	go.uber.org/zap v1.28.0
)

require (
	github.com/ctx42/testing v0.57.1 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace github.com/ctx42/logkit => ../..
