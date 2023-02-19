package fancylog

import (
	"golang.org/x/crypto/ssh/terminal"
)

type HttpLog struct {
	FancyLogger
}

var httpFormatter string = "<-%s->"

type Methods interface {
	Get(a map[string]any, status int)
	Delete(a map[string]any, status int)
	Connect(a map[string]any, status int)
	Head(a map[string]any, status int)
	Options(a map[string]any, status int)
	Post(a map[string]any, status int)
	Put(a map[string]any, status int)
	Trace(a map[string]any, status int)
}

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

func NewHttpLogger(out FdWriter) FancyLogger {
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       out,
		timestamp: true,
		trace:     true,

		nameFormatter: &httpFormatter,
	}
}

func NewHttpLoggerWithError(out FdWriter, err FdWriter) FancyLogger {
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		err:       err,
		timestamp: true,
		trace:     true,

		nameFormatter: &httpFormatter,
	}
}
func NewHttpLoggerWithName(name string, out FdWriter) FancyLogger {
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

		nameFormatter: &httpFormatter,
	}
}

func NewHttpLoggerWithNameAndError(name string, out FdWriter, err FdWriter) FancyLogger {
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

		nameFormatter: &httpFormatter,
	}
}

func (h *HttpLog) Get(a map[string]any, status int) {
	h.outputMap(getPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Delete(a map[string]any, status int) {
	h.outputMap(deletePrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Connect(a map[string]any, status int) {
	h.outputMap(connectPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Head(a map[string]any, status int) {
	h.outputMap(headPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Options(a map[string]any, status int) {
	h.outputMap(optionsPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Post(a map[string]any, status int) {
	h.outputMap(postPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Put(a map[string]any, status int) {
	h.outputMap(putPrefix, a, false, getStatusColor(status))
}

func (h *HttpLog) Trace(a map[string]any, status int) {
	h.outputMap(tracePrefix, a, false, getStatusColor(status))
}

func getStatusColor(status int) *Color {
	if 100 <= status && status >= 199 {
		return &ColorCyan
	} else if 200 <= status && status >= 299 {
		return &ColorGreen
	} else if 300 <= status && status >= 399 {
		return &ColorOrange
	} else if 400 <= status && status >= 499 {
		return &ColorRed
	} else if 500 <= status && status >= 599 {
		return &ColorFatalRed
	}
	return nil
}
