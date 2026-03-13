package infra

import (
	"net/http"

	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

type (
	httpTracedTransportBeforeFn func(req *http.Request, span *tracer.Span)
	httpTracedTransportConfig   struct {
		before httpTracedTransportBeforeFn
	}
)
type HTTPTracedTransportOptionFn func(cfg *httpTracedTransportConfig)

func WithHTTPTracedTransportBefore(before httpTracedTransportBeforeFn) HTTPTracedTransportOptionFn {
	return func(cfg *httpTracedTransportConfig) {
		cfg.before = before
	}
}

func HTTPTracedTransport(rt http.RoundTripper, serviceName string, optionFns ...HTTPTracedTransportOptionFn) http.RoundTripper {
	cfg := &httpTracedTransportConfig{}
	for _, optionFn := range optionFns {
		optionFn(cfg)
	}

	return httptrace.WrapRoundTripper(rt, []httptrace.RoundTripperOption{
		httptrace.WithBefore(func(req *http.Request, span *tracer.Span) {
			span.SetTag(ext.ServiceName, serviceName)
			span.SetTag(ext.SpanType, ext.SpanTypeHTTP)
			span.SetTag(ext.HTTPMethod, req.Method)
			span.SetTag(ext.HTTPURL, req.URL.Path)
			span.SetTag(ext.TargetHost, req.URL.Hostname())
			span.SetTag(ext.HTTPUserAgent, req.UserAgent())
			if cfg.before != nil {
				cfg.before(req, span)
			}
		}),
	}...)
}
