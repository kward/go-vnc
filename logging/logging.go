/*
Package logging provides a small facade over Go's slog with repo-specific helpers.
*/
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
)

// Verbosity levels used throughout the repo; higher is more verbose.
const (
	FlowLevel      = 2
	FnDeclLevel    = 3
	ResultLevel    = 4
	SpamLevel      = 5
	CrazySpamLevel = 6
)

var verbosity atomic.Int32 // if verbosity >= level, V(level) is true

// SetVerbosity sets the global verbosity level (e.g., 0 disables V checks, 5 enables spam-level logs).
func SetVerbosity(v int) { verbosity.Store(int32(v)) }

// V reports whether logs guarded at the given level should be emitted.
func V(level int) bool { return verbosity.Load() >= int32(level) }

var logger atomic.Value // holds *slog.Logger

func init() {
	logger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

// SetLogger overrides the global logger used by this package.
func SetLogger(l *slog.Logger) {
	if l == nil {
		l = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	logger.Store(l)
}

func get() *slog.Logger { return logger.Load().(*slog.Logger) }

// Structured logging helpers (prefer when you have keyvals).
func Info(msg string, attrs ...any)  { get().Info(msg, attrs...) }
func Debug(msg string, attrs ...any) { get().Debug(msg, attrs...) }
func Warn(msg string, attrs ...any)  { get().Warn(msg, attrs...) }
func Error(msg string, attrs ...any) { get().Error(msg, attrs...) }

// Infof logs an informational message.
func Infof(format string, args ...interface{}) { get().Info(fmt.Sprintf(format, args...)) }

// Debugf logs a debug message.
func Debugf(format string, args ...interface{}) { get().Debug(fmt.Sprintf(format, args...)) }

// Warnf logs a warning message.
func Warnf(format string, args ...interface{}) { get().Warn(fmt.Sprintf(format, args...)) }

// Errorf logs an error message.
func Errorf(format string, args ...interface{}) { get().Error(fmt.Sprintf(format, args...)) }

// FnName returns the calling function name, e.g. "SomeFunction()".
func FnName() string {
	pc := make([]uintptr, 10) // At least 1 entry needed.
	runtime.Callers(2, pc)
	name := runtime.FuncForPC(pc[0]).Name()
	return name[strings.LastIndex(name, ".")+1:] + "()"
}

// FnNameWithArgs returns the calling function name with supplied argument values embedded.
func FnNameWithArgs(format string, args ...interface{}) string {
	pc := make([]uintptr, 10) // At least 1 entry needed.
	runtime.Callers(2, pc)
	name := runtime.FuncForPC(pc[0]).Name()
	a := []interface{}{name[strings.LastIndex(name, ".")+1:]}
	a = append(a, args...)
	return fmt.Sprintf("%s("+format+")", a...)
}
