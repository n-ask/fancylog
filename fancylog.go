package fancylog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// FdWriter interface extends existing io.Writer with file descriptor function
// support. This is typically implemented by os.File (like os.Stdout and os.Stderr).
type FdWriter interface {
	io.Writer
	Fd() uintptr
}

// TimestampFunc is a function type used to generate timestamps for log entries.
type TimestampFunc func() (time time.Time, layout string)

// Logger struct defines the underlying storage for a single logger instance.
type Logger struct {
	name           string
	color          bool
	out            FdWriter
	err            FdWriter
	debug          bool
	trace          bool
	timestamp      bool
	timestampColor *Color
	timestampFn    *TimestampFunc
	quiet          bool
	mu             sync.Mutex

	nameFormatter *string
}

// HasColor returns true if color output is enabled for this logger.
func (l *Logger) HasColor() bool {
	return l.color
}

// Level represents the severity level of a log message.
type Level string

func (l Level) toPrefix() []byte {
	return []byte("[" + l + "]")
}

const (
	// Fatal level indicates a critical error that will cause the application to exit.
	Fatal Level = "FATAL"
	// Error level indicates a significant problem that needs attention.
	Error Level = "ERROR"
	// Warn level indicates a potential issue or important situation.
	Warn Level = "WARN"
	// Info level indicates general informational messages.
	Info Level = "INFO"
	// Debug level indicates detailed information for troubleshooting.
	Debug Level = "DEBUG"
	// Trace level indicates even more granular information than Debug, often including stack traces.
	Trace Level = "TRACE"
)

// Prefix struct defines the visual representation of a log level.
// Text will be the prefix text included in the log.
// Color will be the color applied to the log prefix.
// File flag set to true will display code trace (stack trace).
type Prefix struct {
	Text  Level
	Color Color
	File  bool
}

// PrefixText struct holds the values of the prefixes to be used and the tail size
// to add spaces to the end of the prefix for alignment.
type PrefixText struct {
	value    Level
	tailSize int
}

// Prefixes maps each Level to its corresponding Prefix configuration.
var Prefixes = map[Level]Prefix{
	Fatal: {
		Text:  Fatal,
		Color: ColorFatalRed,
		File:  true,
	},
	Error: {
		Text:  Error,
		Color: ColorRed,
	},
	Warn: {
		Text:  Warn,
		Color: ColorOrange,
	},
	Info: {
		Text:  Info,
		Color: ColorGreen,
	},
	Trace: {
		Text:  Trace,
		Color: ColorCyan,
		File:  true,
	},
	Debug: {
		Text:  Debug,
		Color: ColorPurple,
	},
}

// TODO make this a map and get rid of tailSize
var (
	maxNameSize   int = 0
	maxPrefixSize int = 0
)

func scanPrefixes() {
	for level, _ := range Prefixes {
		if l := len(level); l > maxPrefixSize {
			maxPrefixSize = l
		}
	}
}

func defaultTimeFn() (time.Time, string) {
	return time.Now().UTC(), time.RFC3339
}

// New returns a new Logger instance with a predefined writer output.
// It automatically detects if the terminal supports coloring.
func New(out FdWriter) *Logger {
	scanPrefixes()
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       out,
		timestamp: true,
		trace:     true,
	}
}

// NewWithError returns a new Logger instance with separate writers for standard and error output.
// It automatically detects if the terminal supports coloring.
// 'out' would typically be os.Stdout and 'err' would be os.Stderr.
func NewWithError(out FdWriter, err FdWriter) *Logger {
	scanPrefixes()
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       err,
		timestamp: true,
		trace:     true,
	}
}

// NewWithName returns a new Logger instance with a specific name and writer output.
// The name is used as a prefix for all log messages.
func NewWithName(name string, out FdWriter) *Logger {
	scanPrefixes()
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &Logger{
		name:      name,
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       out,
		timestamp: true,
		trace:     true,
	}
}

// NewWithNameAndError returns a new Logger instance with a specific name and separate
// writers for standard and error output.
func NewWithNameAndError(name string, out FdWriter, err FdWriter) *Logger {
	scanPrefixes()
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &Logger{
		name:      name,
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       err,
		timestamp: true,
		trace:     true,
	}
}

// WithColor explicitly enables colorful features on the log.
func (l *Logger) WithColor() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = true
	return l
}

// WithoutColor explicitly disables colorful features on the log.
func (l *Logger) WithoutColor() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = false
	return l
}

// WithDebug enables debugging output to reveal debug and trace level logs.
func (l *Logger) WithDebug() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = true
	return l
}

// WithoutDebug disables debugging output.
func (l *Logger) WithoutDebug() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = false
	return l
}

// WithTrace enables trace output to reveal trace level logs.
func (l *Logger) WithTrace() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.trace = true
	return l
}

// WithoutTrace disables trace output.
func (l *Logger) WithoutTrace() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.trace = false
	return l
}

// IsDebug checks the current state of debugging output.
func (l *Logger) IsDebug() bool {
	return l.debug
}

// IsTrace checks the current state of trace output.
func (l *Logger) IsTrace() bool {
	return l.debug
}

// WithTimestamp enables timestamp output on the log.
func (l *Logger) WithTimestamp() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = true
	return l
}

// WithoutTimestamp disables timestamp output on the log.
func (l *Logger) WithoutTimestamp() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = false
	return l
}

// Quiet disables all log output.
func (l *Logger) Quiet() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = true
	return l
}

// NoQuiet enables all log output (disables quiet mode).
func (l *Logger) NoQuiet() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = false
	return l
}

// IsQuiet checks if the logger is in quiet mode.
func (l *Logger) IsQuiet() bool {
	return l.quiet
}

// SetTimestampColor overrides the default color for timestamps.
func (l *Logger) SetTimestampColor(color Color) {
	l.timestampColor = &color
}

// SetDefaultTimeFn overrides the default timestamp producer.
func (l *Logger) SetDefaultTimeFn(timestampFunc TimestampFunc) {
	l.timestampFn = &timestampFunc
}

func (l *Logger) getTimeFunc() TimestampFunc {
	if l.timestampFn == nil {
		return defaultTimeFn
	}
	return *l.timestampFn
}

func (l *Logger) writePrefix(p Prefix, b ColorLogger, colorOverride *Color) {
	if l.color {
		if colorOverride != nil {
			b.AppendWithColor(p.Text.toPrefix(), *colorOverride)
		} else {
			b.AppendWithColor(p.Text.toPrefix(), p.Color)
		}
	} else {
		b.Append(p.Text.toPrefix())
	}
	for i := 0; i < maxPrefixSize-len(p.Text); i++ {
		b.AppendSpace()
	}
	b.AppendSpace()
}

func (l *Logger) writeTime(b ColorLogger) {
	if l.color {
		if l.timestampColor != nil {
			b.WriteColor(*l.timestampColor)
		} else {
			b.Blue()
		}

	}
	b.AppendTime(l.getTimeFunc()())
	b.AppendSpace()
	// Print reset color if color enabled
	if l.color {
		b.Off()
	}
}

func (l *Logger) writeName(b ColorLogger) {
	if l.color {
		b.NicePurple()
	}
	if l.nameFormatter != nil {
		b.AppendString(fmt.Sprintf(*l.nameFormatter, l.name))
		b.AppendSpace()
	} else {
		b.Append([]byte("<"))
		b.Append([]byte(l.name))
		b.Append([]byte("> "))
	}

	if l.color {
		b.Off()
	}
}

const DepthSkip = 3

// getFile produces the stack trace given the caller and depth
func (l *Logger) getStackTrace() string {
	// Get the caller filename and line
	pcs := make([]uintptr, 20)
	_ = runtime.Callers(2, pcs)
	var stack strings.Builder
	for _, i := range pcs {
		if i != 0 {
			f := runtime.FuncForPC(i)
			if !strings.HasPrefix(strings.ToLower(f.Name()), "github.com/n-ask/fancylog") {
				file, line := f.FileLine(i)
				stack.WriteString(fmt.Sprintf("\t%s()\n\t\t %s:%d\n", f.Name(), file, line))
			}
		}
	}
	return stack.String()
}

func (l *Logger) writeStack(stack string, b ColorLogger) {
	// Print color start if enabled
	if l.color {
		b.Orange()
	}
	// Print filename and line
	b.Append([]byte(stack))
	// Print color stop
	if l.color {
		b.Off()
	}
}

// output print the actual value
func (l *Logger) output(prefix Prefix, data string, isErr bool, prefixColorOverride *Color) {

	// Check if quiet is requested, and try to return no error and be quiet
	if l.IsQuiet() {
		return
	}

	// Temporary storage for file and line tracing
	var stack string

	// Check if the specified prefix needs to be included with file logging
	if prefix.File {
		stack = l.getStackTrace()
	}
	b := NewColorLogger()
	// Reset buffer so it start from the begining
	b.Reset()
	// Write prefix to the buffer
	if len(l.name) > 0 {
		l.writeName(b)
	}
	if len(l.name) != maxNameSize {
		for i := 0; i < (maxNameSize - len(l.name)); i++ {
			b.AppendSpace()
		}
		if len(l.name) == 0 {
			for i := 0; i < 3; i++ {
				b.AppendSpace()
			}
		}
	}

	l.writePrefix(prefix, b, prefixColorOverride)

	// Check if the log require timestamping
	if l.timestamp {
		l.writeTime(b)
	}

	// Print the actual string data from caller
	b.Append([]byte(data))
	if len(data) == 0 || data[len(data)-1] != '\n' {
		b.AppendByte('\n')
	}
	// Add caller filename and line if enabled
	if prefix.File {
		l.writeStack(stack, b)
	}

	if isErr {
		_, _ = l.err.Write(b.Bytes())
	} else {
		_, _ = l.out.Write(b.Bytes())
	}

	b.Free()
	return
}

func (l *Logger) outputMap(prefix Prefix,
	data map[string]interface{},
	isErr bool,
	prefixColorOverride *Color,
	mapKeyColorOverride *map[string]Color,
) {
	// Check if quiet is requested, and try to return no error and be quiet
	if l.IsQuiet() {
		return
	}

	// Temporary storage for file and line tracing
	var stack string

	// Check if the specified prefix needs to be included with file logging
	if prefix.File {
		stack = l.getStackTrace()
	}

	b := NewColorLogger()

	// Reset buffer so it start from the begining
	b.Reset()
	if len(l.name) > 0 {
		l.writeName(b)
	}
	if len(l.name) != maxNameSize {
		for i := 0; i < (maxNameSize - len(l.name)); i++ {
			b.AppendSpace()
		}
		if len(l.name) == 0 {
			for i := 0; i < 3; i++ {
				b.AppendSpace()
			}
		}
	}

	l.writePrefix(prefix, b, prefixColorOverride)

	// Check if the log require timestamping
	// Check if the log require timestamping
	if l.timestamp {
		l.writeTime(b)
	}

	sortedKeys := make([]string, 0, len(data))
	for key := range data {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		if l.color {
			b.Purple()
		}
		if mapKeyColorOverride != nil {
			if c, ok := (*mapKeyColorOverride)[key]; ok {
				b.WriteColor(c)
			}
		}
		b.Append([]byte(key))
		switch t := data[key].(type) {
		case map[string]any:
			b.Append([]byte("["))
			innerSortedKeys := make([]string, 0, len(t))
			for innerKey := range t {
				innerSortedKeys = append(innerSortedKeys, innerKey)
			}
			sort.Strings(innerSortedKeys)
			for _, sortedKey := range innerSortedKeys {
				b.AppendSpace()
				if l.color {
					b.Orange()
				}
				b.Append([]byte(sortedKey))
				if l.color {
					b.White()
				}
				b.Append([]byte(":"))
				if l.color {
					b.Cyan()
				}
				b.Append([]byte(fmt.Sprintf("%+v", t[sortedKey])))
				b.AppendSpace()
			}
			if l.color {
				b.Purple()
			}
			b.Append([]byte("]"))
			b.AppendSpace()
		case map[string][]string:
			b.Append([]byte("["))
			innerSortedKeys := make([]string, 0, len(t))
			for innerKey := range t {
				innerSortedKeys = append(innerSortedKeys, innerKey)
			}
			sort.Strings(innerSortedKeys)
			for _, sortedKey := range innerSortedKeys {
				b.AppendSpace()
				if l.color {
					b.Orange()
				}
				b.Append([]byte(sortedKey))
				if l.color {
					b.White()
				}
				b.Append([]byte(":"))
				if l.color {
					b.Cyan()
				}
				b.Append([]byte(fmt.Sprintf("%+v", t[sortedKey])))
				b.AppendSpace()
			}
			if l.color {
				b.Purple()
			}
			b.Append([]byte("]"))
			b.AppendSpace()
		default:
			if l.color {
				b.Orange()
			}
			b.Append([]byte("="))
			if l.color {
				b.Cyan()
			}
			b.Append([]byte(fmt.Sprintf("%+v", data[key])))
			b.AppendSpace()
		}

	}
	if l.color {
		b.Off()
	}
	b.AppendByte('\n')
	// Add caller filename and line if enabled
	if prefix.File {
		l.writeStack(stack, b)
		b.AppendByte('\n')
	}

	if isErr {
		_, _ = l.err.Write(b.Bytes())
	} else {
		_, _ = l.out.Write(b.Bytes())
	}

	b.Free()
	return
}

// Fatal prints a fatal message to the output and exits the application with status 1.
func (l *Logger) Fatal(v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintln(v...), true, nil)
	os.Exit(1)
}

// FatalWithCode prints a fatal message to the output and exits the application
// with the provided status code.
func (l *Logger) FatalWithCode(exit int, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintln(v...), true, nil)
	os.Exit(exit)
}

// Fatalf prints a formatted fatal message to the output and exits the application with status 1.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintf(format, v...), true, nil)
	os.Exit(1)
}

// FatalWithCodef prints a formatted fatal message to the output and exits the application
// with the provided status code.
func (l *Logger) FatalWithCodef(format string, exit int, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintf(format, v...), true, nil)
	os.Exit(exit)
}

// FatalMap logs structured fatal data and exits the application with status 1.
func (l *Logger) FatalMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Fatal], v, true, nil, nil)
	os.Exit(1)
}

// FatalMapWithCode logs structured fatal data and exits the application with the provided status code.
func (l *Logger) FatalMapWithCode(exit int, v map[string]interface{}) {
	l.outputMap(Prefixes[Fatal], v, true, nil, nil)
	os.Exit(exit)
}

// Error prints an error message to the output.
func (l *Logger) Error(v ...interface{}) {
	l.output(Prefixes[Error], fmt.Sprintln(v...), true, nil)
}

// Errorf prints a formatted error message to the output.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(Prefixes[Error], fmt.Sprintf(format, v...), true, nil)
}

// ErrorMap logs structured error data.
func (l *Logger) ErrorMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Error], v, true, nil, nil)
}

// Warn prints a warning message to the output.
func (l *Logger) Warn(v ...interface{}) {
	l.output(Prefixes[Warn], fmt.Sprintln(v...), false, nil)
}

// Warnf prints a formatted warning message to the output.
func (l *Logger) Warnf(format string, v ...any) {
	l.output(Prefixes[Warn], fmt.Sprintf(format, v...), false, nil)
}

// WarnMap logs structured warning data.
func (l *Logger) WarnMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Warn], v, false, nil, nil)
}

// Info prints an informational message to the output.
func (l *Logger) Info(v ...interface{}) {
	l.output(Prefixes[Info], fmt.Sprintln(v...), false, nil)
}

// Infof prints a formatted informational message to the output.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(Prefixes[Info], fmt.Sprintf(format, v...), false, nil)
}

// InfoMap logs structured informational data.
func (l *Logger) InfoMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Info], v, false, nil, nil)
}

// Debug prints a debug message to the output if debug output is enabled.
func (l *Logger) Debug(v ...interface{}) {
	if l.IsDebug() {
		l.output(Prefixes[Debug], fmt.Sprintln(v...), false, nil)
	}
}

// Debugf prints a formatted debug message to the output if debug output is enabled.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.IsDebug() {
		l.output(Prefixes[Debug], fmt.Sprintf(format, v...), false, nil)
	}
}

// DebugMap logs structured debug data if debug output is enabled.
func (l *Logger) DebugMap(v map[string]interface{}) {
	if l.IsDebug() {
		l.outputMap(Prefixes[Debug], v, false, nil, nil)
	}
}

// Trace prints a trace message to the output if trace output is enabled.
func (l *Logger) Trace(v ...interface{}) {
	if l.IsTrace() {
		l.output(Prefixes[Trace], fmt.Sprintln(v...), false, nil)
	}
}

// Tracef prints a formatted trace message to the output if trace output is enabled.
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.IsTrace() {
		l.output(Prefixes[Trace], fmt.Sprintf(format, v...), false, nil)
	}
}

// TraceMap logs structured trace data if trace output is enabled.
func (l *Logger) TraceMap(v map[string]interface{}) {
	if l.IsTrace() {
		l.outputMap(Prefixes[Trace], v, false, nil, nil)
	}
}

// Log prints a message with a custom prefix.
func (l *Logger) Log(prefix Prefix, a ...any) {
	l.output(prefix, fmt.Sprintln(a...), false, nil)
}

// Logf prints a formatted message with a custom prefix.
func (l *Logger) Logf(prefix Prefix, format string, a ...any) {
	l.output(prefix, fmt.Sprintf(format, a...), false, nil)
}

// LogMap logs structured data with a custom prefix.
func (l *Logger) LogMap(prefix Prefix, a map[string]any) {
	l.outputMap(prefix, a, false, nil, nil)
}
