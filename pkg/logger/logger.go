package logger

import (
	"log"
	"log/slog"
	"net/http"
)

// TODO: Rewrite log to slog.
// TODO: Check what type of logs we can print in this project and оставить разрешенные

// ? TODO: Save logs to the ./logs/triple-s.log path (OPTIONAL)
type iLogger interface {
	PrintfInfoMsg(string, ...interface{})
	PrintfDebugMsg(string, ...interface{})
	PrintfErrorMsg(string, ...interface{})
	LogRequestMiddleware(http.Handler) http.Handler
}

type Logger struct {
	debugMode bool
}

func New(debugMode bool) *Logger {
	return &Logger{
		debugMode: debugMode,
	}
}

func printfMsg(level string, mes string, args ...interface{}) {
	log.Printf(level+" "+mes, args...)
}

func (l *Logger) PrintfInfoMsg(mes string, args ...interface{}) {
	// printfMsg("[INFO]", mes, args...)
	slog.Info(mes, args...)
}

func (l *Logger) PrintfDebugMsg(mes string, args ...interface{}) {
	if l.debugMode {
		printfMsg("[DEBUG]", mes, args...)
		// slog.Debug(mes, args...)
	}
}

func (l *Logger) PrintfErrorMsg(mes string, args ...interface{}) {
	// printfMsg("[ERROR]", mes, args...)
	slog.Error(mes, args...)
}

func (l *Logger) LogRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Request %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
