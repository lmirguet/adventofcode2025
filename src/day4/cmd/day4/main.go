package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"adventofcode2025/day1/src/day4"
)

func main() {
	// -file: ścieżka do pliku zawierającego siatkę. Jeśli nie podano,
	// program czyta ze standardowego wejścia (stdin).
	filePath := flag.String("file", "", "path to grid file (default: stdin)")
	flag.Parse()

	// `reader` będzie albo otwartym plikiem, albo stdin.
	var reader io.ReadCloser
	if *filePath != "" {
		f, err := os.Open(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
			os.Exit(1)
		}
		reader = f
		// zamykamy plik przy wyjściu z main
		defer reader.Close()
	} else {
		reader = os.Stdin
	}

	// Wywołuje logikę oraz obsługuje błędy
	result, err := day4.Compute(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Wyświetla jedynie łączną liczbę usuniętych komórek
	fmt.Fprintf(os.Stdout, "%d\n", result.TotalRemoved)
}
