// Package pgxlog adapts a fancylog.FancyLogger into a pgx query logger.
//
// It is a package of its own so that importing it pulls in the Postgres driver
// and nothing else — and, more to the point, so that callers who want Gin or
// net/http logging do NOT pull it in. These integrations used to share one
// `handlers` package; because Go links at package granularity, importing that
// package for HTTP middleware compiled the pgx file too and put the whole
// Postgres driver stack (eight modules) into binaries that never open a
// database connection.
//
// # pgx v5
//
// This targets pgx v5. v4 is not an option: GO-2026-5004 (SQL injection via
// placeholder confusion with dollar-quoted string literals) affects every v4
// release with no fix available, and v4 drags in github.com/jackc/pgproto3/v2,
// whose GO-2026-4518 denial of service likewise has no fix. v5 carries the
// GO-2026-5004 fix from 5.9.2 onward and vendors its own pgproto3, so both
// findings leave the dependency graph entirely.
//
// The adapter API is otherwise unchanged. pgx v5 removed the pgx.Logger
// interface and moved query logging to pgx/v5/tracelog, whose Logger interface
// takes the same arguments with its own level type. Wire it up with:
//
//	logger := pgxlog.NewFancyPGLogger(myFancyLogger)
//	cfg.Tracer = &tracelog.TraceLog{Logger: logger, LogLevel: tracelog.LogLevelInfo}
package pgxlog

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/n-ask/fancylog"
)

// PgLogger defines the interface for a logger used with pgx.
type PgLogger interface {
	Debug(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Crit(msg string, ctx ...interface{})
}

// IgnoreStmtPrefix defines a SQL statement prefix to be ignored by the logger.
type IgnoreStmtPrefix struct {
	Text       string
	IgnoreCase bool
}

// ContainsPrefix returns true if the given SQL string starts with the defined prefix.
func (i IgnoreStmtPrefix) ContainsPrefix(sql string) bool {
	if i.IgnoreCase {
		return strings.HasPrefix(strings.ToUpper(sql), strings.ToUpper(i.Text))
	}
	return strings.HasPrefix(sql, i.Text)
}

// Compiled once at package level rather than per Log call: these are constant
// patterns, and rebuilding them on every logged query is pure waste in a hot
// path.
var (
	reLeadCloseWhitespace = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	reInsideWhitespace    = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

// FancyPGLogger is an implementation of tracelog.Logger that uses FancyLogger.
type FancyPGLogger struct {
	l       fancylog.FancyLogger
	ignores *[]IgnoreStmtPrefix
}

// Assert at compile time that the adapter still satisfies pgx's logger
// interface, so a future pgx upgrade that changes the signature fails here
// rather than at a call site in some consumer.
var _ tracelog.Logger = (*FancyPGLogger)(nil)

// NewFancyPGLogger creates a new FancyPGLogger instance.
func NewFancyPGLogger(l fancylog.FancyLogger) *FancyPGLogger {
	return &FancyPGLogger{l: l}
}

// SetIgnoreStmtPrefixes configures the logger to ignore SQL statements starting
// with the provided prefixes.
func (l *FancyPGLogger) SetIgnoreStmtPrefixes(p []IgnoreStmtPrefix) {
	l.ignores = &p
}

func (l *FancyPGLogger) containsIgnoredPrefix(sql string) bool {
	if l.ignores != nil {
		for _, prefix := range *l.ignores {
			if prefix.ContainsPrefix(sql) {
				return true
			}
		}
	}
	return false
}

// Log implements the tracelog.Logger interface.
func (l *FancyPGLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	if val, ok := data["sql"]; ok {
		// pgx always supplies a string here, but an unchecked type assertion has
		// no business in a logging path: a malformed entry should log oddly,
		// never panic the process it is only observing.
		if sql, isStr := val.(string); isStr {
			if l.containsIgnoredPrefix(sql) {
				return
			}
			final := reLeadCloseWhitespace.ReplaceAllString(sql, "")
			data["sql"] = reInsideWhitespace.ReplaceAllString(final, " ")
		}
	}
	logArgs := make([]interface{}, 0, len(data)*2)
	for k, v := range data {
		logArgs = append(logArgs, k, v)
	}
	switch level {
	case tracelog.LogLevelTrace:
		l.Debug(msg, append(logArgs, "PGX_LOG_LEVEL", level)...)
	case tracelog.LogLevelDebug:
		l.Debug(msg, logArgs...)
	case tracelog.LogLevelInfo, tracelog.LogLevelNone:
		l.Info(msg, logArgs...)
	case tracelog.LogLevelWarn:
		l.Warn(msg, logArgs...)
	case tracelog.LogLevelError:
		l.Error(msg, logArgs...)
	default:
		l.Error(msg, append(logArgs, "INVALID_PGX_LOG_LEVEL", level)...)
	}
}

// Debug logs a debug message.
func (l *FancyPGLogger) Debug(msg string, ctx ...interface{}) {
	l.l.DebugMap(toMap(msg, ctx))
}

// Info logs an informational message.
func (l *FancyPGLogger) Info(msg string, ctx ...interface{}) {
	l.l.InfoMap(toMap(msg, ctx))
}

// Warn logs a warning message.
func (l *FancyPGLogger) Warn(msg string, ctx ...interface{}) {
	l.l.WarnMap(toMap(msg, ctx))
}

// Error logs an error message.
func (l *FancyPGLogger) Error(msg string, ctx ...interface{}) {
	l.l.ErrorMap(toMap(msg, ctx))
}

// Crit logs a critical (fatal) message.
func (l *FancyPGLogger) Crit(msg string, ctx ...interface{}) {
	l.l.FatalMap(toMap(msg, ctx))
}

func toMap(msg string, ctx []interface{}) map[string]interface{} {
	v := make(map[string]interface{})
	v["msg"] = msg
	if len(ctx)%2 == 0 {
		for i := 0; i < len(ctx); i = i + 2 {
			key, ok := ctx[i].(string)
			if !ok {
				// Same reasoning as the sql assertion above: a non-string key is
				// a caller bug, but it must not panic the logger. Fall back to
				// dumping the raw slice.
				v["data"] = ctx
				return v
			}
			v[key] = ctx[i+1]
		}
	} else {
		v["data"] = ctx
	}
	return v
}
