package fancylog

import (
	"golang.org/x/crypto/ssh/terminal"
)

type FancyHttpLog interface {
	FancyLogger
	Methods

	WithHeaders() FancyHttpLog
	DebugHeaders() bool
}

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

type HttpLog struct {
	FancyLogger
	debugHeaders bool
}

var httpFormatter string = "{%s}"

var (
	getText = &PrefixText{
		value:    "GET",
		tailSize: 0,
	}
	deleteText = &PrefixText{
		value:    "DELETE",
		tailSize: 0,
	}
	connectText = &PrefixText{
		value:    "CONNECT",
		tailSize: 0,
	}
	headText = &PrefixText{
		value:    "HEAD",
		tailSize: 0,
	}
	optionsText = &PrefixText{
		value:    "OPTIONS",
		tailSize: 0,
	}
	postText = &PrefixText{
		value:    "POST",
		tailSize: 0,
	}
	putText = &PrefixText{
		value:    "PUT",
		tailSize: 0,
	}
	traceText = &PrefixText{
		value:    "TRACE",
		tailSize: 0,
	}

	getPrefix = Prefix{
		Text:  getText,
		Color: ColorCyan,
	}

	deletePrefix = Prefix{
		Text:  deleteText,
		Color: nil,
	}

	connectPrefix = Prefix{
		Text:  connectText,
		Color: nil,
	}

	headPrefix = Prefix{
		Text:  headText,
		Color: nil,
	}

	optionsPrefix = Prefix{
		Text:  optionsText,
		Color: nil,
	}

	postPrefix = Prefix{
		Text:  postText,
		Color: nil,
	}

	putPrefix = Prefix{
		Text:  putText,
		Color: nil,
	}

	tracePrefix = Prefix{
		Text:  traceText,
		Color: nil,
	}
)

func NewHttpLogger(out FdWriter) FancyHttpLog {
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

func NewHttpLoggerWithError(out FdWriter, err FdWriter) FancyHttpLog {
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
func NewHttpLoggerWithName(name string, out FdWriter) FancyHttpLog {
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

func NewHttpLoggerWithNameAndError(name string, out FdWriter, err FdWriter) FancyHttpLog {
	if maxNameSize < len(name) {
		maxNameSize = len(name)
	}
	return &HttpLog{
		FancyLogger: &Logger{
			name:          name,
			color:         terminal.IsTerminal(int(out.Fd())),
			out:           out,
			err:           err,
			timestamp:     true,
			trace:         true,
			nameFormatter: &httpFormatter,
		},
		debugHeaders: false,
	}
}

func (h *HttpLog) WithHeaders() FancyHttpLog {
	h.debugHeaders = true
	return h
}

func (h *HttpLog) DebugHeaders() bool {
	return h.debugHeaders
}

func (h *HttpLog) GetMethod(a map[string]any, status int) {
	h.outputMap(getPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) DeleteMethod(a map[string]any, status int) {
	h.outputMap(deletePrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) ConnectMethod(a map[string]any, status int) {
	h.outputMap(connectPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) HeadMethod(a map[string]any, status int) {
	h.outputMap(headPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) OptionsMethod(a map[string]any, status int) {
	h.outputMap(optionsPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) PostMethod(a map[string]any, status int) {
	h.outputMap(postPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) PutMethod(a map[string]any, status int) {
	h.outputMap(putPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) TraceMethod(a map[string]any, status int) {
	h.outputMap(tracePrefix, a, false, getStatusColor(status))
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
