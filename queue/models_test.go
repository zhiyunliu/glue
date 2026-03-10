package queue_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/xtypes"
)

func TestNewMsg(t *testing.T) {

	type test struct {
		OBj1 json.RawMessage `json:"obj1"`
	}

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		obj  interface{}
		opts []queue.MsgOption
		want queue.Message
	}{
		{name: "1.", obj: &test{OBj1: json.RawMessage(`{"a":1}`)}, want: &queue.MsgWrap{HeaderMap: make(xtypes.SMap), BodyBytes: json.RawMessage(`{"obj1":{"a":1}}`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queue.NewMsg(tt.obj, tt.opts...)
			if err != nil {
				t.Errorf("NewMsg() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
