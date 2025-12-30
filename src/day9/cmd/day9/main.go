package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"adventofcode2025/day1/src/day9"
)

// main wires file/stdin input to the day9 solver and prints the final answer.
func main() {
	filePath := flag.String("file", "", "path to tile coordinate file (default: stdin)")
	flag.Parse()

	var reader io.ReadCloser
	if *filePath != "" {
		f, err := os.Open(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
			os.Exit(1)
		}
		reader = f
		defer reader.Close()
	} else {
		reader = os.Stdin
	}

	result, err := day9.Compute(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%d\n", result.MaxArea)
}

