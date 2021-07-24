package libs

import "testing"

func TestMd5(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name  string
		args  args
		wantR string
	}{
		{name: "1", args: args{val: "1"}, wantR: "c4ca4238a0b923820dcc509a6f75849b"},
	}
	for _, tt := range tests {
		//	t.Run(tt.name, func(t *testing.T) {
		if gotR := Md5(tt.args.val); gotR != tt.wantR {
			t.Errorf("Md5() = %v, want %v", gotR, tt.wantR)
		}
		//})
	}
}
