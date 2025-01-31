package fancylog

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// FdWriter interface extends existing io.Writer with file descriptor function
// support
type FdWriter interface {
	io.Writer
	Fd() uintptr
}

type TimestampFunc func() (time time.Time, layout string)

// Logger struct define the underlying storage for single logger
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

func (l *Logger) HasColor() bool {
	return l.color
}

type Level string

func (l Level) toPrefix() []byte {
	return []byte("[" + l + "]")
}

const (
	Fatal Level = "FATAL"
	Error Level = "ERROR"
	Warn  Level = "WARN"
	Info  Level = "INFO"
	Debug Level = "DEBUG"
	Trace Level = "TRACE"
)

// Prefix struct define plain and color byte
// Text will prefix text to include in the log
// Color will be the color applied to the log
// File flag set to true will display code trace
type Prefix struct {
	Text  Level
	Color Color
	File  bool
}

// PrefixText struct to hold the values of the prefixes to be used, and the tail size to add spaces to the end
// of the prefix
type PrefixText struct {
	value    Level
	tailSize int
}

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

// New returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
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

// NewWithError returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support. out would be something like os.Stdout
// and err would be something like os.Stderr
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

// NewWithName {(name string out FdWriter) *Logger { returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
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

// NewWithNameAndError {(name string out FdWriter) *Logger { returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
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

// WithColor explicitly turn on colorful features on the log
func (l *Logger) WithColor() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = true
	return l
}

// WithoutColor explicitly turn off colorful features on the log
func (l *Logger) WithoutColor() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = false
	return l
}

// WithDebug turn on debugging output on the log to reveal debug and trace level
func (l *Logger) WithDebug() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = true
	return l
}

// WithoutDebug turn off debugging output on the log
func (l *Logger) WithoutDebug() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = false
	return l
}

// WithTrace turn on trace output on the log to reveal debug and trace level
func (l *Logger) WithTrace() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.trace = true
	return l
}

// WithoutTrace turn off trace output on the log
func (l *Logger) WithoutTrace() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.trace = false
	return l
}

// IsDebug check the state of debugging output
func (l *Logger) IsDebug() bool {
	return l.debug
}

// IsTrace check the state of trace output
func (l *Logger) IsTrace() bool {
	return l.debug
}

// WithTimestamp turn on timestamp output on the log
func (l *Logger) WithTimestamp() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = true
	return l
}

// WithoutTimestamp turn off timestamp output on the log
func (l *Logger) WithoutTimestamp() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = false
	return l
}

// Quiet turn off all log output
func (l *Logger) Quiet() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = true
	return l
}

// NoQuiet turn on all log output
func (l *Logger) NoQuiet() FancyLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = false
	return l
}

// IsQuiet check for quiet state
func (l *Logger) IsQuiet() bool {
	return l.quiet
}

// SetTimestampColor override the default color for timestamps
func (l *Logger) SetTimestampColor(color Color) {
	l.timestampColor = &color
}

// SetDefaultTimeFn override the default timestamp producer
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

// Fatal print fatal message to output and quit the application with status 1
func (l *Logger) Fatal(v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintln(v...), true, nil)
	os.Exit(1)
}

// FatalWithCode print formatted fatal message to output and quit the application
// with status code provider
func (l *Logger) FatalWithCode(exit int, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintln(v...), true, nil)
	os.Exit(exit)
}

// Fatalf print formatted fatal message to output and quit the application
// with status 1
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintf(format, v...), true, nil)
	os.Exit(1)
}

// FatalWithCodef print formatted fatal message to output and quit the application
// with status code provider
func (l *Logger) FatalWithCodef(format string, exit int, v ...interface{}) {
	l.output(Prefixes[Fatal], fmt.Sprintf(format, v...), true, nil)
	os.Exit(exit)
}

func (l *Logger) FatalMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Fatal], v, true, nil, nil)
	os.Exit(1)
}

func (l *Logger) FatalMapWithCode(exit int, v map[string]interface{}) {
	l.outputMap(Prefixes[Fatal], v, true, nil, nil)
	os.Exit(exit)
}

// Error print error message to output
func (l *Logger) Error(v ...interface{}) {
	l.output(Prefixes[Error], fmt.Sprintln(v...), true, nil)
}

// Errorf print formatted error message to output
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(Prefixes[Error], fmt.Sprintf(format, v...), true, nil)
}

func (l *Logger) ErrorMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Error], v, true, nil, nil)
}

// Warn print warning message to output
func (l *Logger) Warn(v ...interface{}) {
	l.output(Prefixes[Warn], fmt.Sprintln(v...), false, nil)
}

// Warnf print formatted warning message to output
func (l *Logger) Warnf(format string, v ...any) {
	l.output(Prefixes[Warn], fmt.Sprintf(format, v...), false, nil)
}

func (l *Logger) WarnMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Warn], v, false, nil, nil)
}

// Info print informational message to output
func (l *Logger) Info(v ...interface{}) {
	l.output(Prefixes[Info], fmt.Sprintln(v...), false, nil)
}

// Infof print formatted informational message to output
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(Prefixes[Info], fmt.Sprintf(format, v...), false, nil)
}

func (l *Logger) InfoMap(v map[string]interface{}) {
	l.outputMap(Prefixes[Info], v, false, nil, nil)
}

// Debug print debug message to output if debug output enabled
func (l *Logger) Debug(v ...interface{}) {
	if l.IsDebug() {
		l.output(Prefixes[Debug], fmt.Sprintln(v...), false, nil)
	}
}

// Debugf print formatted debug message to output if debug output enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.IsDebug() {
		l.output(Prefixes[Debug], fmt.Sprintf(format, v...), false, nil)
	}
}

func (l *Logger) DebugMap(v map[string]interface{}) {
	if l.IsDebug() {
		l.outputMap(Prefixes[Debug], v, false, nil, nil)
	}
}

// Trace print trace message to output if debug output enabled
func (l *Logger) Trace(v ...interface{}) {
	if l.IsTrace() {
		l.output(Prefixes[Trace], fmt.Sprintln(v...), false, nil)
	}
}

// Tracef print formatted trace message to output if debug output enabled
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.IsTrace() {
		l.output(Prefixes[Trace], fmt.Sprintf(format, v...), false, nil)
	}
}

// TraceMap print formatted trace message to output if debug output enabled
func (l *Logger) TraceMap(v map[string]interface{}) {
	if l.IsTrace() {
		l.outputMap(Prefixes[Trace], v, false, nil, nil)
	}
}

func (l *Logger) Log(prefix Prefix, a ...any) {
	l.output(prefix, fmt.Sprintln(a...), false, nil)
}

func (l *Logger) Logf(prefix Prefix, format string, a ...any) {
	l.output(prefix, fmt.Sprintf(format, a...), false, nil)
}

func (l *Logger) LogMap(prefix Prefix, a map[string]any) {
	l.outputMap(prefix, a, false, nil, nil)
}
