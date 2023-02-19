package fancylog

import (
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/exp/maps"
	"sync"
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
	once         *sync.Once
}

var httplogInit = &sync.Once{}

func httplogInitalizer() {
	maps.Copy(Prefixes, HttpPrefixes)
	scanPrefixes()
}

var httpFormatter string = "{%s}"

const (
	GetLevel     Level = "GET"
	DeleteLevel  Level = "DELETE"
	ConnectLevel Level = "CONNECT"
	HeadLevel    Level = "HEAD"
	OptionsLevel Level = "OPTIONS"
	PostLevel    Level = "POST"
	PutLevel     Level = "PUT"
	TraceLevel   Level = "TRACE"
)

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

func NewHttpLoggerWithNameAndError(name string, out FdWriter, err FdWriter) FancyHttpLog {
	httplogInit.Do(httplogInitalizer)
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

func (h *HttpLog) ensureStatusKey(a map[string]any, status int, prefix Prefix) {
	a["status"] = status
	h.outputMap(prefix, a, false, getStatusColor(status), &map[string]Color{
		"status": ColorOrange,
	})
}

func (h *HttpLog) GetMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[GetLevel])
}

func (h *HttpLog) DeleteMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[DeleteLevel])
}

func (h *HttpLog) ConnectMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[ConnectLevel])
}

func (h *HttpLog) HeadMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[HeadLevel])
}

func (h *HttpLog) OptionsMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[OptionsLevel])
}

func (h *HttpLog) PostMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[PostLevel])
}

func (h *HttpLog) PutMethod(a map[string]any, status int) {
	h.ensureStatusKey(a, status, HttpPrefixes[PutLevel])
}

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
