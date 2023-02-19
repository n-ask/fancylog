package handlers

import (
	"context"
	"github.com/n-ask/fancylog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
)

func logLevelForMap(l fancylog.FancyLogger, level fancylog.Level, val map[string]any) {
	switch level {
	case fancylog.Fatal:
		l.FatalMap(val)
	case fancylog.Error:
		l.ErrorMap(val)
	case fancylog.Warn:
		l.WarnMap(val)
	case fancylog.Debug:
		l.Debug(val)
	case fancylog.Trace:
		l.TraceMap(val)
	case fancylog.Info:
		fallthrough
	default:
		l.InfoMap(val)
	}
}

func UnaryServerInterceptor(l fancylog.FancyLogger, logLevel fancylog.Level, logFunc func(ctx context.Context) map[string]any, ignoreMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if ignored(info.FullMethod, ignoreMethods) {
			return resp, err
		}
		logVal := logFunc(ctx)
		logVal["grpc.method"] = info.FullMethod
		if err != nil {
			logVal["error"] = err.Error()
			l.ErrorMap(logVal)
			return resp, err
		}
		logLevelForMap(l, logLevel, logVal)
		return resp, err
	}
}

func StreamServerInterceptor(l fancylog.FancyLogger, logLevel fancylog.Level, logFunc func(ctx context.Context) map[string]interface{}, ignoreMethods ...string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if ignored(info.FullMethod, ignoreMethods) {
			return err
		}
		logVal := logFunc(stream.Context())
		logVal["grpc.method"] = info.FullMethod
		if err != nil {
			logVal["error"] = err.Error()
			l.ErrorMap(logVal)
			return err
		}
		logLevelForMap(l, logLevel, logVal)
		return err
	}
}

func UnaryClientInterceptor(l fancylog.FancyLogger, logLevel fancylog.Level, logFunc func(ctx context.Context) map[string]interface{}, ignoreMethods ...string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		v := logFunc(ctx)
		if err != nil {
			v["error"] = err.Error()
			l.ErrorMap(v)
			return err
		}
		v["method"] = method
		if ignored(method, ignoreMethods) {
			return err
		}
		logLevelForMap(l, logLevel, v)
		return err
	}
}

func StreamClientInterceptor(l fancylog.FancyLogger, logLevel fancylog.Level, logFunc func(ctx context.Context) map[string]interface{}, ignoreMethods ...string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := streamer(ctx, desc, cc, method, opts...)
		var clientStream *clientStream
		if err == nil {
			clientStream = wrapClientStream(s, desc)
		}
		go func() {
			err := <-clientStream.finished
			v := logFunc(ctx)
			if err != nil {
				v["error"] = err.Error()
				l.ErrorMap(v)
				return
			}
			v["method"] = method
			if ignored(method, ignoreMethods) {
				return
			}
			logLevelForMap(l, logLevel, v)
		}()
		return clientStream, err
	}
}

func ignored(method string, methods []string) bool {
	if methods == nil {
		return false
	}
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func wrapClientStream(s grpc.ClientStream, desc *grpc.StreamDesc) *clientStream {
	events := make(chan streamEvent)
	eventsDone := make(chan struct{})
	finished := make(chan error)
	go func() {
		defer close(eventsDone)

		// Both streams have to be closed
		state := byte(0)

		for event := range events {
			switch event.Type {
			case closeEvent:
				state |= clientClosedState
			case receiveEndEvent:
				state |= receiveEndedState
			case errorEvent:
				finished <- event.Err
				return
			}

			if state == clientClosedState|receiveEndedState {
				finished <- nil
				return
			}
		}
	}()

	return &clientStream{
		ClientStream: s,
		desc:         desc,
		events:       events,
		eventsDone:   eventsDone,
		finished:     finished,
	}
}

type clientStream struct {
	grpc.ClientStream

	desc       *grpc.StreamDesc
	events     chan streamEvent
	eventsDone chan struct{}
	finished   chan error

	receivedMessageID int
	sentMessageID     int
}

func (w *clientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	if err == nil && !w.desc.ServerStreams {
		w.sendStreamEvent(receiveEndEvent, nil)
	} else if err == io.EOF {
		w.sendStreamEvent(receiveEndEvent, nil)
	} else if err != nil {
		w.sendStreamEvent(errorEvent, err)
	} else {
		w.receivedMessageID++
		//messageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}
func (w *clientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)

	w.sentMessageID++
	//messageSent.Event(w.Context(), w.sentMessageID, m)

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return err
}
func (w *clientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}
	return md, err
}
func (w *clientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	} else {
		w.sendStreamEvent(closeEvent, nil)
	}

	return err
}
func (w *clientStream) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-w.eventsDone:
	case w.events <- streamEvent{Type: eventType, Err: err}:
	}
}

type streamEventType int

const (
	closeEvent streamEventType = iota
	receiveEndEvent
	errorEvent
)

type streamEvent struct {
	Type streamEventType
	Err  error
}

const (
	clientClosedState byte = 1 << iota
	receiveEndedState
)
