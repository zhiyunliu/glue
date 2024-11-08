package xdb

import "testing"

func TestIsNil(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{name: "0.", input: nil, want: true},
		{name: "1.", input: (*int)(nil), want: true},
		{name: "2.", input: 1, want: false},
		{name: "3.", input: 1.0, want: false},
		{name: "4.", input: struct{ a int }{a: 1}, want: false},
		{name: "5.", input: struct{ a *int }{}, want: false},
		{name: "6.", input: map[string]interface{}{}, want: false},
		{name: "7.", input: []string{}, want: false},
		{name: "8.", input: []string{"a"}, want: false},
		{name: "9.", input: []int{}, want: false},
		{name: "10.", input: []int{1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.input); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
