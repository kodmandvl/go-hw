package main

import (
	"runtime"
	"testing"
)

func TestRunCmd(t *testing.T) {
	env := Environment{
		"FOO": {Value: "bar", NeedRemove: false},
	}

	cmd := shellCommand("echo $FOO")

	code := RunCmd(cmd, env)
	if code != 0 {
		t.Fatalf("expected exit code 0, result: %d", code)
	}
}

func TestRunCmd_RemoveEnv(t *testing.T) {
	t.Setenv("FOO", "should_be_removed")

	env := Environment{
		"FOO": {NeedRemove: true},
	}

	cmd := shellCommand("echo $FOO")

	code := RunCmd(cmd, env)
	if code != 0 {
		t.Fatalf("expected exit code 0, result: %d", code)
	}
}

func TestRunCmd_ExitCodeForward(t *testing.T) {
	cmd := shellCommand("exit 28")

	code := RunCmd(cmd, Environment{})
	if code != 28 {
		t.Fatalf("expected exit code 28, result: %d", code)
	}
}

func shellCommand(script string) []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd.exe", "/C", script}
	}
	return []string{"sh", "-c", script}
}
