package handlers

import (
	"bufio"
	"github.com/felixge/httpsnoop"
	"github.com/n-ask/fancylog"
	"io"
	"net"
	"net/http"
)

type loggingHandler struct {
	writer  io.Writer
	handler http.Handler
	log     *fancylog.Logger
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := l.w.(http.Hijacker).Hijack()
	if err == nil && l.status == 0 {
		// The status will be StatusSwitchingProtocols if there was no error and
		// WriteHeader has not been called yet
		l.status = http.StatusSwitchingProtocols
	}
	return conn, rw, err
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger, w := makeLogger(w)
	url := *r.URL

	h.handler.ServeHTTP(w, r)
	if r.MultipartForm != nil {
		r.MultipartForm.RemoveAll()
	}
	msg := map[string]any{}
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			msg["user"] = name
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
		msg["host"] = host
	}
	msg["uri"] = r.RequestURI
	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if r.ProtoMajor == 2 && r.Method == "CONNECT" {
		msg["uri"] = r.Host
	}
	if msg["uri"] == "" {
		msg["uri"] = url.RequestURI()
	}
	msg["method"] = r.Method
	msg["proto"] = r.Proto
	msg["status"] = logger.Status()
	msg["size"] = logger.Size()

	h.log.InfoMap(msg)
}

func makeLogger(w http.ResponseWriter) (*responseLogger, http.ResponseWriter) {
	logger := &responseLogger{w: w, status: http.StatusOK}
	return logger, httpsnoop.Wrap(w, httpsnoop.Hooks{
		Write: func(httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return logger.Write
		},
		WriteHeader: func(httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return logger.WriteHeader
		},
	})
}

func LoggingHandler(log *fancylog.Logger, out io.Writer, h http.Handler) http.Handler {
	return loggingHandler{
		writer:  out,
		handler: h,
		log:     log,
	}
}
