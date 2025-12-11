package main

import (
	"flag"
	"fmt"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

// Example:
// go run . -from in.log -to out.log -offset 0 -limit 0
// cmp out.log in.log
// go run . -from in.log -to out.log -offset 10 -limit 20

func main() {
	flag.Parse()

	if from == "" || to == "" {
		fmt.Println("Usage of hw07_file_copying: -from <src> -to <dest> [-offset <num>] [-limit <num>]")
		return
	}

	if err := Copy(from, to, offset, limit); err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Copied from", from, "to", to, "( offset", offset, ", limit", limit, "): [OK]")
}
