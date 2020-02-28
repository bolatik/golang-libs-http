package http

type Logger interface {
	Info(args ...interface{})

	Warn(args ...interface{})

	Fatal(args ...interface{})

	Debug(args ...interface{})
}
