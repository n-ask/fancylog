package handlers

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/n-ask/fancylog"
	"regexp"
	"strings"
)

type PgLogger interface {
	Debug(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Crit(msg string, ctx ...interface{})
}

type Ignore struct {
}

type IgnoreStmtPrefix struct {
	Text       string
	IgnoreCase bool
}

func (i IgnoreStmtPrefix) ContainsPrefix(sql string) bool {
	if i.IgnoreCase {
		return strings.HasPrefix(strings.ToUpper(sql), strings.ToUpper(i.Text))
	}
	return strings.HasPrefix(sql, i.Text)
}

type FancyPGLogger struct {
	l       *fancylog.Logger
	ignores *[]IgnoreStmtPrefix
}

func NewFancyPGLogger(l *fancylog.Logger) *FancyPGLogger {
	return &FancyPGLogger{l: l}
}

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

func (l *FancyPGLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if val, ok := data["sql"]; ok {
		sql := val.(string)
		re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
		re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
		final := re_leadclose_whtsp.ReplaceAllString(sql, "")
		final = re_inside_whtsp.ReplaceAllString(final, " ")
		if l.containsIgnoredPrefix(sql) {
			//Eat
			return
		}
		data["sql"] = final
	}
	logArgs := make([]interface{}, 0, len(data))
	for k, v := range data {
		logArgs = append(logArgs, k, v)
	}
	switch level {
	case pgx.LogLevelTrace:
		l.Debug(msg, append(logArgs, "PGX_LOG_LEVEL", level)...)
	case pgx.LogLevelDebug:
		l.Debug(msg, logArgs...)
	case pgx.LogLevelInfo, pgx.LogLevelNone:
		l.Info(msg, logArgs...)
	case pgx.LogLevelWarn:
		l.Warn(msg, logArgs...)
	case pgx.LogLevelError:
		l.Error(msg, logArgs...)
	default:
		l.Error(msg, append(logArgs, "INVALID_PGX_LOG_LEVEL", level)...)
	}
}

func (l *FancyPGLogger) Debug(msg string, ctx ...interface{}) {
	l.l.DebugMap(toMap(msg, ctx))
}

func (l *FancyPGLogger) Info(msg string, ctx ...interface{}) {
	l.l.InfoMap(toMap(msg, ctx))
}

func (l *FancyPGLogger) Warn(msg string, ctx ...interface{}) {
	l.l.WarnMap(toMap(msg, ctx))
}

func (l *FancyPGLogger) Error(msg string, ctx ...interface{}) {
	l.l.ErrorMap(toMap(msg, ctx))
}

func (l *FancyPGLogger) Crit(msg string, ctx ...interface{}) {
	l.l.FatalMap(toMap(msg, ctx))
}

func toMap(msg string, ctx []interface{}) map[string]interface{} {
	v := make(map[string]interface{})
	v["msg"] = msg
	if len(ctx)%2 == 0 {
		for i := 0; i < len(ctx); i = i + 2 {
			v[ctx[i].(string)] = ctx[i+1]
		}
	} else {
		v["data"] = ctx
	}
	return v
}
