package testutil

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/sobadon/anrd/internal/errutil"
)

func TestErrorsAs(t *testing.T) {
	type args struct {
		err    error
		target interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil と nil は true",
			args: args{
				err:    nil,
				target: nil,
			},
			want: true,
		},
		{
			name: "nil と error は false",
			args: args{
				err:    nil,
				target: errutil.ErrInternal,
			},
			want: false,
		},
		{
			name: "error と nil は false",
			args: args{
				err:    errutil.ErrInternal,
				target: nil,
			},
			want: false,
		},
		{
			name: "同一の error と error は true（wrap なし）",
			args: args{
				err:    errutil.ErrInternal,
				target: errutil.ErrInternal,
			},
			want: true,
		},
		{
			name: "同一の error と error は true（wrap あり）",
			args: args{
				err:    errors.Wrap(errutil.ErrInternal, "something happen"),
				target: errutil.ErrInternal,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorsAs(tt.args.err, tt.args.target); got != tt.want {
				t.Errorf("ErrorsAs() = %v, want %v", got, tt.want)
			}
		})
	}
}
