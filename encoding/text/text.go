package text

import (
	"fmt"

	"github.com/zhiyunliu/gel/encoding"
	"github.com/zhiyunliu/golibs/bytesconv"
)

const Name = "text"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with json.
type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	str, _ := v.(string)
	return bytesconv.StringToBytes(str), nil

}

func (codec) Unmarshal(data []byte, v interface{}) error {
	str, ok := v.(*string)
	if !ok {
		return fmt.Errorf("text type error,%s", data)
	}

	*str = bytesconv.BytesToString(data)
	return nil
}

func (codec) Name() string {
	return Name
}
