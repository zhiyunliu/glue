package registry

import (
	"fmt"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/golibs/xnet"
	"github.com/zhiyunliu/golibs/xsecurity/aes"
)

var _cache cmap.ConcurrentMap
var _factoryMap = map[string]Factory{}

func init() {
	_cache = cmap.New()
}

//IFactory 注册中心构建器
type Factory interface {
	Name() string
	Create(config.Config) (Registrar, error)
}

//Register 添加注册中心工厂对象
func Register(factory Factory) {
	if factory == nil {
		panic(fmt.Errorf("registry: factory is nil"))
	}
	name := factory.Name()
	if _, ok := _factoryMap[name]; ok {
		panic(fmt.Errorf("registry: factory called twice:" + name))
	}
	_factoryMap[name] = factory
}

func GetRegistrar(cfg config.Config) (Registrar, error) {
	//nacos://xxxx
	cfgVal := cfg.Value("registry").String()

	tmpVal := _cache.Upsert(cfgVal, nil, func(exist bool, valueInMap, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}

		protoName, configName, err := xnet.Parse(cfgVal)
		if err != nil {
			return err
		}
		factory, ok := _factoryMap[protoName]
		if !ok {
			return fmt.Errorf("registrar不支持的Proto类型[%s]", protoName)
		}
		regCfg := cfg.Get(protoName).Get(configName)
		bval, err := regCfg.Value("encrypt").Bool()
		if err != nil {
			return fmt.Errorf("registrar:%s://%s.encrypt 配置有误：%w", protoName, configName, err)
		}
		if bval {
			edval := regCfg.Value("data").String()
			plainText, err := aes.Decrypt(edval, cfg.Value("app.encrypt").String(), "cbc/pkcs8")
			if err != nil {
				return fmt.Errorf("registrar: 解密失败：%w Data:%s", err, edval)
			}
			regCfg = config.New(config.WithSource(config.NewStrSource(plainText)))
		}
		tmpv, err := factory.Create(regCfg)
		if err != nil {
			return err
		}
		return tmpv

	})

	if err, ok := tmpVal.(error); ok {
		return nil, err
	}

	return tmpVal.(Registrar), nil

}
