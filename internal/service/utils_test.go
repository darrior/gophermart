package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getPasswordHash(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Valid test",
			args: args{
				password: "1234",
			},
			want: "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getPasswordHash(tt.args.password))
		})
	}
}
