package xgrpc

const RoundRobin = "round_robin"

//Option 配置选项
type Option func(*setting)

//WithConnectionTimeout 配置网络连接超时时长
func WithConnectionTimeout(seconds int) Option {
	return func(o *setting) {
		if seconds < 0 {
			return
		}
		o.ConntTimeout = seconds
	}
}

//WithBalancer 配置为负载均衡器
func WithBalancer(balancer string) Option {
	return func(o *setting) {
		o.Balancer = balancer
	}
}

//WithRoundRobin 配置为轮询负载均衡器
func WithRoundRobin() Option {
	return func(o *setting) {
		o.Balancer = RoundRobin
	}
}

type RequestOption func(*requestOptions)

type requestOptions struct {
	Header       map[string]string
	TraceId      string
	WaitForReady bool
}

func WithHeaders(header map[string]string) RequestOption {
	return func(o *requestOptions) {
		o.Header = header
	}
}

func WithTraceID(traceId string) RequestOption {
	return func(o *requestOptions) {
		o.TraceId = traceId
	}
}
func WithWaitForReady(waitForReady bool) RequestOption {
	return func(o *requestOptions) {
		o.WaitForReady = waitForReady
	}
}
