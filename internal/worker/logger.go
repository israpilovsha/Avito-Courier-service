package worker

type Logger interface {
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}
