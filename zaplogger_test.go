package hs

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewZapLogger(t *testing.T) {
	l, err := NewZapLogger(LogConf{
		"debug",
		[]string{"stdout", "/tmp/test_stdout_1.log", "/tmp/test_stdout_2.log"},
		[]string{"stderr", "/tmp/test_stderr_1.log", "/tmp/test_stderr_2.log"},
	},
	)
	defer l.Sync()
	require.NoError(t, err)
	l.Info("this is info log")
	l.Error("this is error log")
	l.Sugar().Info("this is sugar info")
}
