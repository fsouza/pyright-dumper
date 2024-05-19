package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	output := flag.String("output", "", "output path")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	if *output == "" {
		flag.Usage()
		os.Exit(2)
	}

	pyright := "pyright"
	if p, ok := os.LookupEnv("PYRIGHT"); ok {
		pyright = p
	}

	cmd := exec.Command(pyright, append([]string{"--watch"}, flag.Args()...)...)
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = consume(out, *output, *verbose)
	if err != nil {
		log.Fatal(err)
	}
}

func consume(r io.Reader, outputFile string, verbose bool) error {
	const marker = "Watching for file changes..."

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if verbose {
			log.Print(line)
		}
		lines = append(lines, line)
		if line == marker {
			err := write(lines, outputFile)
			if err != nil {
				return err
			}
			lines = lines[:0]
		}
	}

	return scanner.Err()
}

func write(lines []string, outputFile string) error {
	dir := filepath.Dir(outputFile)
	tmpfile, err := os.CreateTemp(dir, "")
	if err != nil {
		return err
	}
	defer tmpfile.Close()

	for _, line := range lines {
		fmt.Fprintln(tmpfile, line)
	}

	return os.Rename(tmpfile.Name(), outputFile)
}
