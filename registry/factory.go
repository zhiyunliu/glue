package registry

import (
	"fmt"

	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/golibs/xnet"
	"github.com/zhiyunliu/golibs/xsecurity/aes"
)

//IFactory 注册中心构建器
type Factory interface {
	Name() string
	Create(config.Config) (Registrar, error)
}

var factoryMap = map[string]Factory{}

//Register 添加注册中心工厂对象
func Register(factory Factory) {
	if factory == nil {
		panic(fmt.Errorf("registry: factory is nil"))
	}
	name := factory.Name()
	if _, ok := factoryMap[name]; ok {
		panic(fmt.Errorf("registry: factory called twice:" + name))
	}
	factoryMap[name] = factory
}

func GetRegistrar(cfg config.Config) (Registrar, error) {
	//nacos://xxxx
	cfgVal := cfg.Value("registry").String()

	protoName, configName, err := xnet.Parse(cfgVal)
	if err != nil {
		return nil, err
	}
	factory, ok := factoryMap[protoName]
	if !ok {
		return nil, fmt.Errorf("Registrar不支持的Proto类型[%s]", protoName)
	}
	regCfg := cfg.Get(protoName).Get(configName)
	bval, err := regCfg.Value("encrypt").Bool()
	if err != nil {
		return nil, fmt.Errorf("Registrar:%s://%s.encrypt 配置有误：%+v", protoName, configName, err)
	}
	if bval {
		edval := regCfg.Value("data").String()
		plainText, err := aes.Decrypt(edval, cfg.Value("app.encrypt").String(), "cbc/pkcs8")
		if err != nil {
			return nil, fmt.Errorf("Registrar: 解密失败：%+v \nData:%s", err, edval)
		}
		regCfg = config.New(config.WithSource(config.NewStrSource(plainText)))
	}

	return factory.Create(regCfg)
}
