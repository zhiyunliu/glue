package errors_test

import (
	"fmt"
	"testing"

	"github.com/zhiyunliu/glue/errors"
)

func TestCode(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		err  error
		want int
	}{
		//	{name: "bad request", err: errors.BadRequest("message"), want: 400},
		//	{name: "Unauthorized", err: errors.Unauthorized("message"), want: 401},
		{name: "wraperror", err: fmt.Errorf("wrap:%w", errors.Unauthorized("message")), want: 401},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Code(tt.err)
			// TODO: update the condition below to compare got with tt.want.
			if tt.want != got {
				t.Errorf("Code() = %v, want %v", got, tt.want)
			}
		})
	}
}
