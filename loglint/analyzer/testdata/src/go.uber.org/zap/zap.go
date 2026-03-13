package zap

type Field struct{}

type Logger struct{}

func NewExample() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string, fields ...Field) {}

func Int(key string, value int) Field   { return Field{} }
func String(key, value string) Field    { return Field{} }

