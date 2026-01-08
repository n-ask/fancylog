package fancylog

import (
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/exp/maps"
	"sync"
)

// FancyHttpLog is an extension of FancyLogger specialized for HTTP request logging.
// It includes methods for logging various HTTP methods with status codes.
type FancyHttpLog interface {
	FancyLogger
	Methods

	// WithHeaders enables logging of HTTP headers.
	WithHeaders() FancyHttpLog
	// DebugHeaders returns true if header logging is enabled.
	DebugHeaders() bool
}

// Methods defines the logging methods for different HTTP verbs.
type Methods interface {
	GetMethod(a map[string]any, status int)
	DeleteMethod(a map[string]any, status int)
	ConnectMethod(a map[string]any, status int)
	HeadMethod(a map[string]any, status int)
	OptionsMethod(a map[string]any, status int)
	PostMethod(a map[string]any, status int)
	PutMethod(a map[string]any, status int)
	TraceMethod(a map[string]any, status int)
}

// HttpLog implements the FancyHttpLog interface.
type HttpLog struct {
	FancyLogger
	debugHeaders bool
	once         *sync.Once
}

var httplogInit = &sync.Once{}

func httplogInitalizer() {
	maps.Copy(Prefixes, HttpPrefixes)
	scanPrefixes()
}

var httpFormatter string = "{%s}"

const (
	// GetLevel represents the GET HTTP method.
	GetLevel Level = "GET"
	// DeleteLevel represents the DELETE HTTP method.
	DeleteLevel Level = "DELETE"
	// ConnectLevel represents the CONNECT HTTP method.
	ConnectLevel Level = "CONNECT"
	// HeadLevel represents the HEAD HTTP method.
	HeadLevel Level = "HEAD"
	// OptionsLevel represents the OPTIONS HTTP method.
	OptionsLevel Level = "OPTIONS"
	// PostLevel represents the POST HTTP method.
	PostLevel Level = "POST"
	// PutLevel represents the PUT HTTP method.
	PutLevel Level = "PUT"
	// TraceLevel represents the TRACE HTTP method.
	TraceLevel Level = "TRACE"
)

// HttpPrefixes maps HTTP levels to their corresponding Prefix configuration.
var HttpPrefixes = map[Level]Prefix{
	GetLevel: {
		Text:  GetLevel,
		Color: ColorCyan,
	},
	DeleteLevel: {
		Text:  DeleteLevel,
		Color: ColorCyan,
	},
	ConnectLevel: {
		Text:  ConnectLevel,
		Color: ColorCyan,
	},
	HeadLevel: {
		Text:  HeadLevel,
		Color: ColorCyan,
	},
	OptionsLevel: {
		Text:  OptionsLevel,
		Color: ColorCyan,
	},
	PostLevel: {
		Text:  PostLevel,
		Color: ColorCyan,
	},
	PutLevel: {
		Text:  PutLevel,
		Color: ColorCyan,
	},
	TraceLevel: {
		Text:  TraceLevel,
		Color: ColorCyan,
	},
}

// NewHttpLogger returns a new HttpLog instance with a predefined writer output.
func NewHttpLogger(out FdWriter) FancyHttpLog {
	httplogInit.Do(httplogInitalizer)
	return &HttpLog{
		FancyLogger: &Logger{
			color:         terminal.IsTerminal(int(out.Fd())),
			out:           out,
			err:           out,
			timestamp:     true,
			trace:         true,
			nameFormatter: &httpFormatter,
		},
		debugHeaders: false,
	}
}

// NewHttpLoggerWithError returns a new HttpLog instance with separate writers for
// standard and error output.
func NewHttpLoggerWithError(out FdWriter, err FdWriter) FancyHttpLog {
	httplogInit.Do(httplogInitalizer)
	return &HttpLog{
		FancyLogger: &Logger{
			color:     terminal.IsTerminal(int(out.Fd())),
			out:       out,
			err:       err,
			timestamp: true,
			trace:     true,

			nameFormatter: &httpFormatter,
		},
		debugHeaders: false,
	}
}

// NewHttpLoggerWithName returns a new HttpLog instance with a specific name and writer output.
func NewHttpLoggerWithName(name string, out FdWriter) FancyHttpLog {
	httplogInit.Do(httplogInitalizer)
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &HttpLog{
		FancyLogger: &Logger{
			name:      name,
			color:     terminal.IsTerminal(int(out.Fd())),
			out:       out,
			err:       out,
			timestamp: true,
			trace:     true,

			nameFormatter: &httpFormatter,
		},
		debugHeaders: false,
	}
}

// NewHttpLoggerWithNameAndError returns a new HttpLog instance with a specific name
// and separate writers for standard and error output.
func NewHttpLoggerWithNameAndError(name string, out FdWriter, err FdWriter) FancyHttpLog {
	httplogInit.Do(httplogInitalizer)
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &HttpLog{
		FancyLogger: &Logger{
			name:          name,
			color:         true,
			out:           out,
			err:           err,
			timestamp:     true,
			trace:         true,
			nameFormatter: &httpFormatter,
		},
		debugHeaders: false,
	}
}

// WithHeaders enables header logging for this HttpLog instance.
func (h *HttpLog) WithHeaders() FancyHttpLog {
	h.debugHeaders = true
	return h
}

// DebugHeaders returns whether header logging is enabled.
func (h *HttpLog) DebugHeaders() bool {
	return h.debugHeaders
}

func (h *HttpLog) ensureStatusKey(a map[string]any, status int, prefix Prefix) {
	a["status"] = status

	h.outputMap(prefix, a, false, getStatusColor(status), &map[string]Color{
		"status": ColorOrange,
	})
}

// GetMethod logs a GET request with structured data and status code.
func (h *HttpLog) GetMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[GetLevel])
}

// DeleteMethod logs a DELETE request with structured data and status code.
func (h *HttpLog) DeleteMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[DeleteLevel])
}

// ConnectMethod logs a CONNECT request with structured data and status code.
func (h *HttpLog) ConnectMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[ConnectLevel])
}

// HeadMethod logs a HEAD request with structured data and status code.
func (h *HttpLog) HeadMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[HeadLevel])
}

// OptionsMethod logs an OPTIONS request with structured data and status code.
func (h *HttpLog) OptionsMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[OptionsLevel])
}

// PostMethod logs a POST request with structured data and status code.
func (h *HttpLog) PostMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[PostLevel])
}

// PutMethod logs a PUT request with structured data and status code.
func (h *HttpLog) PutMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[PutLevel])
}

// TraceMethod logs a TRACE request with structured data and status code.
func (h *HttpLog) TraceMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[TraceLevel])
}

func getStatusColor(status int) *Color {
	if 100 <= status && status <= 199 {
		return &ColorCyan
	} else if 200 <= status && status <= 299 {
		return &ColorGreen
	} else if 300 <= status && status <= 399 {
		return &ColorOrange
	} else if 400 <= status && status <= 499 {
		return &ColorRed
	} else if 500 <= status && status <= 599 {
		return &ColorFatalRed
	}
	return nil
}
