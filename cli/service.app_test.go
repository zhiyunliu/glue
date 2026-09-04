package cli

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/config/file"
)

func TestServiceApp_deduplicateConfigSource(t *testing.T) {
	opts := &Options{}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfgsource []string
		want      []string
	}{
		{name: "1.", cfgsource: []string{"a.json", "b.json", "a.json", "c.json", "c.json", "a.json"}, want: []string{"b.json", "c.json", "a.json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &ServiceApp{
				appName:        tt.name,
				options:        opts,
				closeWaitGroup: &sync.WaitGroup{},
			}

			cfgsource := make([]config.Source, 0, len(tt.cfgsource))
			for _, src := range tt.cfgsource {
				cfgsource = append(cfgsource, file.NewSource(src))
			}

			got := app.deduplicateConfigSource(cfgsource)

			gotStr := make([]string, 0, len(got))
			for _, src := range got {
				gotStr = append(gotStr, src.Path())
			}

			// TODO: update the condition below to compare got with tt.want.
			if !reflect.DeepEqual(gotStr, tt.want) {
				t.Errorf("deduplicateConfigSource() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
