package engine

import "fmt"

type LogLevel int

const (
	LevelDebug LogLevel = 1 << iota
	LevelInfo
	LevelWarn
	LevelError

	LevelNone = 0
	LevelAll  = LevelDebug | LevelInfo | LevelWarn | LevelError
)

var logLevel LogLevel = LevelAll

func AddLogLevel(l LogLevel) {
	logLevel |= l
}

func RemoveLogLevel(l LogLevel) {
	logLevel &= ^l
}

func LogDebug(f string, args ...any) {
	if logLevel&LevelDebug == 0 {
		return
	}

	fmt.Println("debug:", fmt.Sprintf(f, args...))
}

func LogInfo(f string, args ...any) {
	if logLevel&LevelInfo == 0 {
		return
	}

	fmt.Println("info:", fmt.Sprintf(f, args...))
}

func LogWarn(f string, args ...any) {
	if logLevel&LevelWarn == 0 {
		return
	}

	fmt.Println("warn:", fmt.Sprintf(f, args...))
}

func LogError(f string, args ...any) {
	if logLevel&LevelError == 0 {
		return
	}

	fmt.Println("error:", fmt.Sprintf(f, args...))
}
