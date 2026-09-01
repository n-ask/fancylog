// Package ginlog provides Gin middleware that logs each request through a
// fancylog.FancyHttpLog.
//
// It is a package of its own so that importing it pulls in Gin and nothing
// else. These integrations used to share one `handlers` package, which meant
// a caller wanting Gin middleware also linked gRPC and the Postgres driver.
package ginlog

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/n-ask/fancylog"
)

// GinLogger is a middleware function for the Gin framework that logs each request
// using the provided FancyHttpLog instance. It captures details like client IP,
// request URI, protocol, status code, and optionally headers.
func GinLogger(logger fancylog.FancyHttpLog) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()
		url := *c.Request.URL

		msg := map[string]any{}
		if url.User != nil {
			if name := url.User.Username(); name != "" {
				msg["user"] = name
			}
		}
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			host = c.Request.RemoteAddr
			msg["host"] = host
		}
		msg["uri"] = c.Request.RequestURI
		// Requests using the CONNECT method over HTTP/2.0 must use
		// the authority field (aka r.Host) to identify the target.
		// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
		if c.Request.ProtoMajor == 2 && c.Request.Method == "CONNECT" {
			msg["uri"] = c.Request.Host
		}
		if msg["uri"] == "" {
			msg["uri"] = url.RequestURI()
		}
		msg["clientIp"] = c.ClientIP()
		msg["remoteIp"] = c.RemoteIP()
		//msg["method"] = r.Method
		msg["proto"] = c.Request.Proto
		msg["status"] = c.Writer.Status()
		msg["size"] = c.Writer.Size()

		if logger.DebugHeaders() {
			headers := map[string][]string{}
			for header, v := range c.Request.Header {
				headers[header] = v
			}
			msg["headers"] = headers
		}

		switch c.Request.Method {
		case http.MethodGet:
			logger.GetMethod(msg, c.Writer.Status())
		case http.MethodConnect:
			logger.ConnectMethod(msg, c.Writer.Status())
		case http.MethodDelete:
			logger.DeleteMethod(msg, c.Writer.Status())
		case http.MethodHead:
			logger.HeadMethod(msg, c.Writer.Status())
		case http.MethodOptions:
			logger.OptionsMethod(msg, c.Writer.Status())
		case http.MethodPost:
			logger.PostMethod(msg, c.Writer.Status())
		case http.MethodPut:
			logger.PutMethod(msg, c.Writer.Status())
		case http.MethodTrace:
			logger.TraceMethod(msg, c.Writer.Status())
		default:
			msg["method"] = c.Request.Method
			logger.InfoMap(msg)
		}
	}
}
