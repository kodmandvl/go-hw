package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		desired Environment
		wantErr bool
	}{
		{
			name: "envdir_is_correct",
			dir:  "testdata/env",
			desired: Environment{
				"HELLO": EnvValue{Value: `"hello"`, NeedRemove: false}, // именно так, в кавычках
				"FOO":   EnvValue{Value: "   foo\nwith new line", NeedRemove: false},
				"BAR":   EnvValue{Value: "bar", NeedRemove: false},
				"UNSET": EnvValue{Value: "", NeedRemove: true},  // файл пустой
				"EMPTY": EnvValue{Value: "", NeedRemove: false}, // в файе есть одна строка, но пустая
			},
			wantErr: false,
		},
		{
			name:    "dir_does_not_exist",
			dir:     "testdata/failed_dir",
			desired: Environment{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ReadDir(tt.dir)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.desired, res)
		})
	}
}
