package infra

import (
	"net/http"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type beforeFn func(req *http.Request, span ddtrace.Span)
type config struct {
	before beforeFn
}
type OptionFn func(c *config)

func WithBefore(before beforeFn) OptionFn {
	return func(c *config) {
		c.before = before
	}
}

func HTTPTracedTransport(rt http.RoundTripper, serviceName string, optionFns ...OptionFn) http.RoundTripper {
	cfg := &config{}
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
