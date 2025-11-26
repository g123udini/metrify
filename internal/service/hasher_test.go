package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignData(t *testing.T) {
	type arg struct {
		value []byte
		key   string
	}

	tests := []struct {
		name string
		arg1 arg
		arg2 arg
		want bool
	}{
		{
			name: "SignData same data, same key",
			arg1: arg{value: []byte("hello world"), key: "test"},
			arg2: arg{value: []byte("hello world"), key: "test"},
			want: true,
		},
		{
			name: "SignData same data, diff key",
			arg1: arg{value: []byte("hello world"), key: "no"},
			arg2: arg{value: []byte("hello world"), key: "test"},
			want: false,
		},
		{
			name: "SignData diff data, diff key",
			arg1: arg{value: []byte("no"), key: "no"},
			arg2: arg{value: []byte("hello world"), key: "test"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := SignData(tt.arg1.value, tt.arg1.key)
			got2 := SignData(tt.arg2.value, tt.arg2.key)

			equal := got1 == got2

			assert.Equal(t, tt.want, equal, "case %q", tt.name)
		})
	}
}
