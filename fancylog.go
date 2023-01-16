package fancylog

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	Text  *prefixText
	Color Color
	File  bool
}

// prefixText struct to hold the values of the prefixes to be used, and the tail size to add spaces to the end
// of the prefix
type prefixText struct {
	value    Level
	tailSize int
}

var (
	plainFatal = &prefixText{
		value:    Fatal,
		tailSize: 0,
	}
	plainError = &prefixText{
		value:    Error,
		tailSize: 0,
	}
	plainWarn = &prefixText{
		value:    Warn,
		tailSize: 1,
	}
	plainInfo = &prefixText{
		value:    Info,
		tailSize: 1,
	}
	plainDebug = &prefixText{
		value:    Debug,
		tailSize: 0,
	}
	plainTrace = &prefixText{
		value:    Trace,
		tailSize: 0,
	}

	// FatalPrefix show fatal prefix
	FatalPrefix = Prefix{
		Text:  plainFatal,
		Color: ColorFatalRed,
		File:  true,
	}

	// ErrorPrefix show error prefix
	ErrorPrefix = Prefix{
		Text:  plainError,
		Color: ColorRed,
		File:  true,
	}

	// WarnPrefix show warn prefix
	WarnPrefix = Prefix{
		Text:  plainWarn,
		Color: ColorOrange,
	}

	// InfoPrefix show info prefix
	InfoPrefix = Prefix{
		Text:  plainInfo,
		Color: ColorGreen,
	}

	// TracePrefix show info prefix
	TracePrefix = Prefix{
		Text:  plainTrace,
		Color: ColorCyan,
	}

	// DebugPrefix show info prefix
	DebugPrefix = Prefix{
		Text:  plainDebug,
		Color: ColorPurple,
		File:  true,
	}

	maxNameSize int = 0
)

func defaultTimeFn() (time.Time, string) {
	return time.Now(), time.RFC3339
}

// New returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
func New(out FdWriter) *Logger {
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       out,
		timestamp: true,
	}
}

// NewWithError returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support. out would be something like os.Stdout
// and err would be something like os.Stderr
func NewWithError(out FdWriter, err FdWriter) *Logger {
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       err,
		timestamp: true,
	}
}

// NewWithName {(name string out FdWriter) *Logger { returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
func NewWithName(name string, out FdWriter) *Logger {
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &Logger{
		name:      name,
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       out,
		timestamp: true,
	}
}

// NewWithNameAndError {(name string out FdWriter) *Logger { returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
func NewWithNameAndError(name string, out FdWriter, err FdWriter) *Logger {
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &Logger{
		name:      name,
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       err,
		timestamp: true,
	}
}

// WithColor explicitly turn on colorful features on the log
func (l *Logger) WithColor() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = true
	return l
}

// WithoutColor explicitly turn off colorful features on the log
func (l *Logger) WithoutColor() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = false
	return l
}

// WithDebug turn on debugging output on the log to reveal debug and trace level
func (l *Logger) WithDebug() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = true
	return l
}

// WithoutDebug turn off debugging output on the log
func (l *Logger) WithoutDebug() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = false
	return l
}

// WithTrace turn on trace output on the log to reveal debug and trace level
func (l *Logger) WithTrace() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.trace = true
	return l
}

// WithoutTrace turn off trace output on the log
func (l *Logger) WithoutTrace() *Logger {
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
func (l *Logger) WithTimestamp() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = true
	return l
}

// WithoutTimestamp turn off timestamp output on the log
func (l *Logger) WithoutTimestamp() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = false
	return l
}

// Quiet turn off all log output
func (l *Logger) Quiet() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = true
	return l
}

// NoQuiet turn on all log output
func (l *Logger) NoQuiet() *Logger {
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

func (l *Logger) writePrefix(p Prefix, b ColorLogger) {
	if p.Text != nil {
		if l.color {
			b.AppendWithColor(p.Text.value.toPrefix(), p.Color)
		} else {
			b.Append(p.Text.value.toPrefix())
		}
		for i := 0; i < p.Text.tailSize; i++ {
			b.AppendSpace()
		}
		b.AppendSpace()
	}
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
	b.Append([]byte("<"))
	b.Append([]byte(l.name))
	b.Append([]byte("> "))
	if l.color {
		b.Off()
	}
}

const DepthSkip = 3

// getFile produces the stack trace given the caller and depth
func (l *Logger) getFile() (file, fn string, line int) {
	var ok bool
	var pc uintptr
	// Get the caller filename and line
	if pc, file, line, ok = runtime.Caller(DepthSkip); !ok {
		file = "<unknown file>"
		fn = "<unknown function>"
		line = 0
	} else {
		file = filepath.Base(file)
		fn = runtime.FuncForPC(pc).Name()
	}
	return
}

func (l *Logger) writeStack(file, fn string, line int, b ColorLogger) {
	// Print color start if enabled
	if l.color {
		b.Orange()
	}
	// Print filename and line
	b.Append([]byte(fn))
	b.AppendByte(':')
	b.Append([]byte(file))
	b.AppendByte(':')
	b.AppendInt(int64(line))
	b.AppendByte(' ')
	// Print color stop
	if l.color {
		b.Off()
	}
}

// output print the actual value
func (l *Logger) output(prefix Prefix, data string, isErr bool) {
	// Check if quiet is requested, and try to return no error and be quiet
	if l.IsQuiet() {
		return
	}

	// Temporary storage for file and line tracing
	var file string
	var line int
	var fn string

	// Check if the specified prefix needs to be included with file logging
	if prefix.File {
		file, fn, line = l.getFile()
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

	l.writePrefix(prefix, b)

	// Check if the log require timestamping
	if l.timestamp {
		l.writeTime(b)
	}
	// Add caller filename and line if enabled
	if prefix.File {
		l.writeStack(file, fn, line, b)
	}

	// Print the actual string data from caller
	b.Append([]byte(data))
	if len(data) == 0 || data[len(data)-1] != '\n' {
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

func (l *Logger) outputMap(prefix Prefix, data map[string]interface{}, isErr bool) {
	// Check if quiet is requested, and try to return no error and be quiet
	if l.IsQuiet() {
		return
	}

	// Temporary storage for file and line tracing
	var file string
	var line int
	var fn string

	// Check if the specified prefix needs to be included with file logging
	if prefix.File {
		file, fn, line = l.getFile()
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

	l.writePrefix(prefix, b)

	// Check if the log require timestamping
	// Check if the log require timestamping
	if l.timestamp {
		l.writeTime(b)
	}

	// Add caller filename and line if enabled
	if prefix.File {
		l.writeStack(file, fn, line, b)
	}

	for key, val := range data {
		if l.color {
			b.Purple()
		}
		b.Append([]byte(key))
		if l.color {
			b.Orange()
		}
		b.Append([]byte("="))
		if l.color {
			b.Cyan()
		}
		b.Append([]byte(fmt.Sprintf("%+v", val)))
		b.AppendSpace()
		if l.color {
			b.Off()
		}
	}
	b.AppendByte('\n')

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
	l.output(FatalPrefix, fmt.Sprintln(v...), true)
	os.Exit(1)
}

// FatalWithCode print formatted fatal message to output and quit the application
// with status code provider
func (l *Logger) FatalWithCode(exit int, v ...interface{}) {
	l.output(FatalPrefix, fmt.Sprintln(v...), true)
	os.Exit(exit)
}

// Fatalf print formatted fatal message to output and quit the application
// with status 1
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(FatalPrefix, fmt.Sprintf(format, v...), true)
	os.Exit(1)
}

// FatalWithCodef print formatted fatal message to output and quit the application
// with status code provider
func (l *Logger) FatalWithCodef(format string, exit int, v ...interface{}) {
	l.output(FatalPrefix, fmt.Sprintf(format, v...), true)
	os.Exit(exit)
}

func (l *Logger) FatalMap(v map[string]interface{}) {
	l.outputMap(FatalPrefix, v, true)
	os.Exit(1)
}

func (l *Logger) FatalMapWithCode(exit int, v map[string]interface{}) {
	l.outputMap(FatalPrefix, v, true)
	os.Exit(exit)
}

// Error print error message to output
func (l *Logger) Error(v ...interface{}) {
	l.output(ErrorPrefix, fmt.Sprintln(v...), true)
}

// Errorf print formatted error message to output
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.output(ErrorPrefix, fmt.Sprintf(format, v...), true)
}

func (l *Logger) ErrorMap(v map[string]interface{}) {
	l.outputMap(ErrorPrefix, v, true)
}

// Warn print warning message to output
func (l *Logger) Warn(v ...interface{}) {
	l.output(WarnPrefix, fmt.Sprintln(v...), false)
}

// Warnf print formatted warning message to output
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.output(WarnPrefix, fmt.Sprintf(format, v...), false)
}

func (l *Logger) WarnMap(v map[string]interface{}) {
	l.outputMap(WarnPrefix, v, false)
}

// Info print informational message to output
func (l *Logger) Info(v ...interface{}) {
	l.output(InfoPrefix, fmt.Sprintln(v...), false)
}

// Infof print formatted informational message to output
func (l *Logger) Infof(format string, v ...interface{}) {
	l.output(InfoPrefix, fmt.Sprintf(format, v...), false)
}

func (l *Logger) InfoMap(v map[string]interface{}) {
	l.outputMap(InfoPrefix, v, false)
}

// Debug print debug message to output if debug output enabled
func (l *Logger) Debug(v ...interface{}) {
	if l.IsDebug() {
		l.output(DebugPrefix, fmt.Sprintln(v...), false)
	}
}

// Debugf print formatted debug message to output if debug output enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.IsDebug() {
		l.output(DebugPrefix, fmt.Sprintf(format, v...), false)
	}
}

func (l *Logger) DebugMap(v map[string]interface{}) {
	if l.IsDebug() {
		l.outputMap(DebugPrefix, v, false)
	}
}

// Trace print trace message to output if debug output enabled
func (l *Logger) Trace(v ...interface{}) {
	if l.IsTrace() {
		l.output(TracePrefix, fmt.Sprintln(v...), false)
	}
}

// Tracef print formatted trace message to output if debug output enabled
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.IsTrace() {
		l.output(TracePrefix, fmt.Sprintf(format, v...), false)
	}
}

// TraceMap print formatted trace message to output if debug output enabled
func (l *Logger) TraceMap(v map[string]interface{}) {
	if l.IsTrace() {
		l.outputMap(TracePrefix, v, false)
	}
}
