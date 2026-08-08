package api

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

/*
```

	"api":{
		"config":{"addr":":8080","engine":"gin","http_protocol":"h2c","h2c":{},"h2":{"cert_file":"","key_file":""},"status":"start/stop","read_timeout":10,"write_timeout":10,"read_header_timeout":10,"max_header_bytes":65525},
		"middlewares":[
		{
			"auth":{
				"proto":"jwt",
				"jwt":{},
				"exclude":["/**"]
			}
		},{}],
		"header":{},
	}

```
*/
type serverConfig struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
	Header      engine.Header       `json:"header"  yaml:"header"`
}

type Config struct {
	Addr               string        `json:"addr"`
	Engine             string        `json:"engine"`
	HttpProtocol       string        `json:"http_protocol" yaml:"http_protocol"`
	H2C                H2cConfig     `json:"h2c" yaml:"h2c"`
	H2                 H2Config      `json:"h2" yaml:"h2"`
	Status             engine.Status `json:"status"`
	ReadTimeout        uint          `json:"read_timeout"`
	WriteTimeout       uint          `json:"write_timeout"`
	StopMaximumTimeout uint          `json:"stop_maximum_timeout"`
	ReadHeaderTimeout  uint          `json:"read_header_timeout"`
	MaxHeaderBytes     uint          `json:"max_header_bytes"`
}

type H2cConfig struct {
	Http2Config `json:",inline" yaml:",inline"`
}

type H2Config struct {
	Http2Config `json:",inline" yaml:",inline"`
	CertFile    string `json:"cert_file" yaml:"cert_file"`
	KeyFile     string `json:"key_file" yaml:"key_file"`
}

type Http2Config struct {
	MaxConcurrentStreams         uint32 `json:"max_concurrent_streams" yaml:"max_concurrent_streams"`
	MaxDecoderHeaderTableSize    uint32 `json:"max_decoder_header_table_size" yaml:"max_decoder_header_table_size"`
	MaxEncoderHeaderTableSize    uint32 `json:"max_encoder_header_table_size" yaml:"max_encoder_header_table_size"`
	MaxReadFrameSize             uint32 `json:"max_read_frame_size" yaml:"max_read_frame_size"`
	MaxUploadBufferPerConnection int32  `json:"max_upload_buffer_per_connection" yaml:"max_upload_buffer_per_connection"`
	MaxUploadBufferPerStream     int32  `json:"max_upload_buffer_per_stream" yaml:"max_upload_buffer_per_stream"`
	IdleTimeout                  uint   `json:"idle_timeout" yaml:"idle_timeout"`
	ReadIdleTimeout              uint   `json:"read_idle_timeout" yaml:"read_idle_timeout"`
	PingTimeout                  uint   `json:"ping_timeout" yaml:"ping_timeout"`
	WriteByteTimeout             uint   `json:"write_byte_timeout" yaml:"write_byte_timeout"`
}
