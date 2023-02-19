package fancylog

type FancyLogger interface {
	StandardLog
	FormatLog
	MappedLog
	PrefixLog

	WithColor() FancyLogger
	WithoutColor() FancyLogger
	WithDebug() FancyLogger
	WithoutDebug() FancyLogger
	WithTrace() FancyLogger
	WithoutTrace() FancyLogger
	IsDebug() bool
	IsTrace() bool
	WithTimestamp() FancyLogger
	WithoutTimestamp() FancyLogger
	Quiet() FancyLogger
	NoQuiet() FancyLogger
	IsQuiet() bool

	output(prefix Prefix, data string, isErr bool, prefixColorOverride *Color)
	outputMap(prefix Prefix, data map[string]interface{}, isErr bool, prefixColorOverride *Color)
}

type StandardLog interface {
	Info(a ...any)
	Debug(a ...any)
	Warn(a ...any)
	Error(a ...any)
	Trace(a ...any)
	Fatal(a ...any)
}

type FormatLog interface {
	Infof(format string, a ...any)
	Debugf(format string, a ...any)
	Warnf(format string, a ...any)
	Errorf(format string, a ...any)
	Tracef(format string, a ...any)
	Fatalf(format string, a ...any)
}

type MappedLog interface {
	InfoMap(a map[string]any)
	DebugMap(a map[string]any)
	WarnMap(a map[string]any)
	ErrorMap(a map[string]any)
	TraceMap(a map[string]any)
	FatalMap(a map[string]any)
}

type PrefixLog interface {
	Log(prefix Prefix, a ...any)
	Logf(prefix Prefix, format string, a ...any)
	LogMap(prefix Prefix, a map[string]any)
}
