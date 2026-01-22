package main

import (
	"fmt"
	"os"
)

/*
Example 1 (testdata/echo.sh):
go run . testdata/env testdata/echo.sh -a 1 -b 2 -c 3.
Example 2 (/usr/local/bin/kubectl):
mkdir -pv /tmp/env
echo '/path/to/my/custom/kubeconfig' > /tmp/env/KUBECONFIG
echo 'nano' > /tmp/env/EDITOR
go run . /tmp/env /usr/local/bin/kubectl edit deploy/mydeploy && echo OK
*/

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage of hw08_envdir_tool: %s <envdir> <command> [<args>]\n", os.Args[0])
		os.Exit(1)
	}

	envdir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(envdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error by ReadDir: %v\n", err)
		os.Exit(1)
	}

	os.Exit(RunCmd(cmd, env))
}
