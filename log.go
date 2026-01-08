// Package fancylog provides stylized, non-blocking log output.
// It is designed to be fast and visually appealing with support for colors
// and structured logging via maps.
package fancylog

// FancyLogger is the primary interface for the fancylogger.
// It combines standard logging, formatted logging, mapped logging, and prefixed logging.
// It also provides configuration methods for colors, debug/trace levels, and quiet mode.
type FancyLogger interface {
	StandardLog
	FormatLog
	MappedLog
	PrefixLog

	// WithColor explicitly enables colorful features on the log.
	WithColor() FancyLogger
	// WithoutColor explicitly disables colorful features on the log.
	WithoutColor() FancyLogger
	// WithDebug enables debugging output to reveal debug and trace level logs.
	WithDebug() FancyLogger
	// WithoutDebug disables debugging output.
	WithoutDebug() FancyLogger
	// WithTrace enables trace output to reveal trace level logs.
	WithTrace() FancyLogger
	// WithoutTrace disables trace output.
	WithoutTrace() FancyLogger
	// IsDebug checks the current state of debugging output.
	IsDebug() bool
	// IsTrace checks the current state of trace output.
	IsTrace() bool
	// WithTimestamp enables timestamp output on the log.
	WithTimestamp() FancyLogger
	// WithoutTimestamp disables timestamp output on the log.
	WithoutTimestamp() FancyLogger
	// Quiet disables all log output.
	Quiet() FancyLogger
	// NoQuiet enables all log output (disables quiet mode).
	NoQuiet() FancyLogger
	// IsQuiet checks if the logger is in quiet mode.
	IsQuiet() bool
	// HasColor checks if the logger has color enabled.
	HasColor() bool

	// output is an internal method to print the actual value.
	output(prefix Prefix, data string, isErr bool, prefixColorOverride *Color)
	// outputMap is an internal method to print mapped data.
	outputMap(prefix Prefix, data map[string]interface{}, isErr bool, prefixColorOverride *Color, mapKeyColorOverride *map[string]Color)
}

// StandardLog defines the standard logging methods for different severity levels.
type StandardLog interface {
	Info(a ...any)
	Debug(a ...any)
	Warn(a ...any)
	Error(a ...any)
	Trace(a ...any)
	Fatal(a ...any)
}

// FormatLog defines the formatted logging methods (printf-style) for different severity levels.
type FormatLog interface {
	Infof(format string, a ...any)
	Debugf(format string, a ...any)
	Warnf(format string, a ...any)
	Errorf(format string, a ...any)
	Tracef(format string, a ...any)
	Fatalf(format string, a ...any)
}

// MappedLog defines the methods for logging structured data using maps.
type MappedLog interface {
	InfoMap(a map[string]any)
	DebugMap(a map[string]any)
	WarnMap(a map[string]any)
	ErrorMap(a map[string]any)
	TraceMap(a map[string]any)
	FatalMap(a map[string]any)
}

// PrefixLog defines methods for logging with custom prefixes.
type PrefixLog interface {
	Log(prefix Prefix, a ...any)
	Logf(prefix Prefix, format string, a ...any)
	LogMap(prefix Prefix, a map[string]any)
}
