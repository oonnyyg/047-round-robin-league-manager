package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"

	"example.com/round-robin-league-manager/internal/league"
)

type request struct {
	Op           string   `json:"op"`
	Teams        []string `json:"teams"`
	MatchID      string   `json:"match_id"`
	HomeGoals    *int     `json:"home_goals"`
	AwayGoals    *int     `json:"away_goals"`
	ThroughRound *int     `json:"through_round"`
}

type matchView struct {
	ID        string `json:"id"`
	Round     int    `json:"round"`
	Home      string `json:"home"`
	Away      string `json:"away"`
	Played    bool   `json:"played"`
	HomeGoals *int   `json:"home_goals,omitempty"`
	AwayGoals *int   `json:"away_goals,omitempty"`
}

type response struct {
	OK        bool         `json:"ok"`
	Error     string       `json:"error,omitempty"`
	Teams     []string     `json:"teams,omitempty"`
	Matches   []matchView  `json:"matches,omitempty"`
	Standings []league.Row `json:"standings,omitempty"`
}

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

func run(stateDir string, in io.Reader, out io.Writer) error {
	var l *league.League
	if loaded, err := league.Load(stateDir); err == nil {
		l = loaded
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load state: %w", err)
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			writeResponse(writer, response{Error: "invalid JSON: " + err.Error()})
			continue
		}
		resp := handle(&l, stateDir, req)
		writeResponse(writer, resp)
	}
	return scanner.Err()
}

func handle(l **league.League, stateDir string, req request) response {
	switch req.Op {
	case "init":
		return handleInit(l, stateDir, req)
	case "schedule":
		return handleSchedule(*l)
	case "register", "correct", "revoke":
		return handleResult(*l, stateDir, req)
	case "standings":
		return handleStandings(*l, req)
	default:
		return response{Error: "unknown op: " + req.Op}
	}
}

func handleInit(l **league.League, stateDir string, req request) response {
	if *l != nil {
		return response{Error: "league already initialized; re-init is not allowed"}
	}
	created, err := league.New(req.Teams)
	if err != nil {
		return response{Error: err.Error()}
	}
	if err := created.Save(stateDir); err != nil {
		return response{Error: "save state: " + err.Error()}
	}
	*l = created
	return response{OK: true, Teams: created.Teams}
}

func handleSchedule(l *league.League) response {
	if l == nil {
		return response{Error: "league is not initialized"}
	}
	views := make([]matchView, len(l.Matches))
	for i, m := range l.Matches {
		v := matchView{ID: m.ID, Round: m.Round, Home: m.Home, Away: m.Away}
		if res, ok := l.Results[m.ID]; ok {
			v.Played = true
			h, a := res.HomeGoals, res.AwayGoals
			v.HomeGoals = &h
			v.AwayGoals = &a
		}
		views[i] = v
	}
	return response{OK: true, Teams: l.Teams, Matches: views}
}

func handleResult(l *league.League, stateDir string, req request) response {
	if l == nil {
		return response{Error: "league is not initialized"}
	}
	if req.Op != "revoke" {
		if req.HomeGoals == nil || req.AwayGoals == nil {
			return response{Error: "home_goals and away_goals are required"}
		}
	}

	var err error
	switch req.Op {
	case "register":
		err = l.Register(req.MatchID, *req.HomeGoals, *req.AwayGoals)
	case "correct":
		err = l.Correct(req.MatchID, *req.HomeGoals, *req.AwayGoals)
	case "revoke":
		err = l.Revoke(req.MatchID)
	}
	if err != nil {
		return response{Error: err.Error()}
	}
	if err := l.Save(stateDir); err != nil {
		return response{Error: "save state: " + err.Error()}
	}
	return response{OK: true}
}

func handleStandings(l *league.League, req request) response {
	if l == nil {
		return response{Error: "league is not initialized"}
	}
	through := 0
	if req.ThroughRound != nil {
		through = *req.ThroughRound
	}
	rows, err := l.Standings(through)
	if err != nil {
		return response{Error: err.Error()}
	}
	return response{OK: true, Standings: rows}
}

func writeResponse(w *bufio.Writer, resp response) {
	if resp.Error != "" {
		resp.OK = false
	} else {
		resp.OK = true
	}
	data, err := json.Marshal(resp)
	if err != nil {
		data = []byte(`{"ok":false,"error":"internal encoding failure"}`)
	}
	w.Write(data)
	w.WriteByte('\n')
	w.Flush()
}
