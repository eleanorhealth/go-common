package infra

import (
	"net/http"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type httpTracedTransportBeforeFn func(req *http.Request, span ddtrace.Span)
type httpTracedTransportConfig struct {
	before httpTracedTransportBeforeFn
}
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
		httptrace.WithBefore(func(req *http.Request, span ddtrace.Span) {
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
