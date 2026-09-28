package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/round-robin-league-manager/internal/league"
)

func main() {
	stateDir := flag.String("state", "", "data directory for league state (required)")
	flag.Parse()
	if *stateDir == "" {
		fmt.Fprintln(os.Stderr, "usage: league --state <data-dir>")
		os.Exit(2)
	}
	if err := run(*stateDir, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dir string, in io.Reader, out io.Writer) error {
	state, err := league.Load(dir)
	if err != nil && err != league.ErrNotInitialized {
		return err
	}
	if err == league.ErrNotInitialized {
		state = nil
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req league.Request
		if err := json.Unmarshal(line, &req); err != nil {
			writeResponse(writer, league.Response{OK: false, Error: "invalid JSON: " + err.Error()})
			continue
		}
		next, resp := league.Handle(state, dir, req)
		state = next
		writeResponse(writer, resp)
	}
	return scanner.Err()
}

func writeResponse(w *bufio.Writer, resp league.Response) {
	data, err := league.EncodeResponse(resp)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\":false,\"error\":\"%s\"}\n", err.Error())
		return
	}
	w.Write(data)
	w.Flush()
}
