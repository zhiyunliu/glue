package nacos

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/zhiyunliu/glue/config"
)

type ClientConfig struct {
	TimeoutMs            uint64 `json:"timeoutms,omitempty"`               // timeout for requesting Nacos server, default value is 10000ms
	BeatInterval         int64  `json:"beatinterval,omitempty"`            // the time interval for sending beat to server,default value is 5000ms
	NamespaceId          string `json:"namespace_id,omitempty"`            // the namespaceId of Nacos.When namespace is public, fill in the blank string here.
	AppName              string `json:"appname,omitempty"`                 // the appName
	Endpoint             string `json:"endpoint,omitempty"`                // the endpoint for get Nacos server addresses
	RegionId             string `json:"regionid,omitempty"`                // the regionId for kms
	AccessKey            string `json:"access_key,omitempty"`              // the AccessKey for kms
	SecretKey            string `json:"secret_key,omitempty"`              // the SecretKey for kms
	CacheDir             string `json:"cache_dir,omitempty"`               // the directory for persist nacos service info,default value is current path
	UpdateCacheWhenEmpty bool   `json:"update_cache_when_empty,omitempty"` // update cache when get empty service instance from server
	Username             string `json:"username,omitempty"`                // the username for nacos auth
	Password             string `json:"password,omitempty"`                // the password for nacos auth
	LogDir               string `json:"log_dir,omitempty"`                 // the directory for log, default is current path
	ContextPath          string `json:"contextpath,omitempty"`             // the nacos server contextpath
}

type ServerConfig struct {
	Scheme      string `json:"scheme,omitempty"`      //the nacos server scheme
	ContextPath string `json:"contextpath,omitempty"` //the nacos server contextpath
	IpAddr      string `json:"ipaddr,omitempty"`      //the nacos server address
	Port        uint64 `json:"port,omitempty"`        //the nacos server port
}

func GetClientParam(cfg config.Config) (param *vo.NacosClientParam, err error) {
	clientConfig := &ClientConfig{
		LogDir:   "../logs/nacos",
		CacheDir: "../conf/nacos",
	}
	serverConfigs := []ServerConfig{}

	err = cfg.Value("client").Scan(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("nacos ClientConfig Scan:%+v", err)
	}
	err = cfg.Value("server").Scan(&serverConfigs)
	if err != nil {
		return nil, fmt.Errorf("nacos ServerConfig Scan:%+v", err)
	}

	ncc := &constant.ClientConfig{}

	cbuffer := &bytes.Buffer{}
	encodeErr := gob.NewEncoder(cbuffer).Encode(clientConfig)
	if encodeErr != nil {
		return nil, fmt.Errorf("nacos ClientConfig Encode:%+v", err)
	}
	decodeErr := gob.NewDecoder(cbuffer).Decode(ncc)
	if decodeErr != nil {
		return nil, fmt.Errorf("nacos ClientConfig Decode:%+v", err)
	}

	nsc := []constant.ServerConfig{}

	sbuffer := &bytes.Buffer{}
	encodeErr = gob.NewEncoder(sbuffer).Encode(serverConfigs)
	if encodeErr != nil {
		return nil, fmt.Errorf("nacos ServerConfig Encode:%+v", err)
	}
	decodeErr = gob.NewDecoder(sbuffer).Decode(&nsc)
	if decodeErr != nil {
		return nil, fmt.Errorf("nacos ServerConfig Decode:%+v", err)
	}

	param = &vo.NacosClientParam{
		ClientConfig:  ncc,
		ServerConfigs: nsc,
	}

	return
}
