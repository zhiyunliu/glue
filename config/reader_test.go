package config

import "testing"

func Test_unmarshalJSON(t *testing.T) {

	type S struct {
		A int  `json:"a"`
		B bool `json:"b"`
	}

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		data    []byte
		v       interface{}
		wantErr bool
		want    S
	}{
		{name: "1.", data: []byte(`{"a":1}`), v: &S{}, wantErr: false, want: S{A: 1}},
		{name: "2.", data: []byte(`{"b":false}`), v: &S{B: true}, wantErr: false, want: S{B: false}},
		{name: "3.", data: []byte(`{}`), v: &S{B: true}, wantErr: false, want: S{B: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := unmarshalJSON(tt.data, tt.v)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("unmarshalJSON() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("unmarshalJSON() succeeded unexpectedly")
			}
			if gotS, ok := tt.v.(*S); !ok {
				t.Fatalf("unexpected type: %T", tt.v)
			} else if *gotS != tt.want {
				t.Errorf("unmarshalJSON() = %+v, want %+v", *gotS, tt.want)
			}
		})
	}
}
