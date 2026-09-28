package league

import (
	"testing"
)

func TestProtocolFullLifecycle(t *testing.T) {
	dir := t.TempDir()
	var state *State

	// Init with out-of-order IDs; response must carry sorted teams.
	state, resp := Handle(state, dir, Request{Action: "init", Teams: []string{"C", "A", "B", "D"}})
	if !resp.OK {
		t.Fatalf("init: %s", resp.Error)
	}
	wantTeams := []string{"A", "B", "C", "D"}
	for i := range wantTeams {
		if resp.Teams[i] != wantTeams[i] {
			t.Fatalf("teams = %v, want %v", resp.Teams, wantTeams)
		}
	}
	if resp.TotalRounds != 3 {
		t.Fatalf("total rounds = %d, want 3", resp.TotalRounds)
	}

	// Second init must be rejected and leave the league intact.
	_, resp = Handle(state, dir, Request{Action: "init", Teams: []string{"X", "Y"}})
	if resp.OK {
		t.Fatal("re-init should be rejected")
	}

	// Schedule: sorted by round then match id, unplayed at first.
	_, resp = Handle(state, dir, Request{Action: "schedule"})
	if !resp.OK || len(resp.Matches) != 6 {
		t.Fatalf("schedule: ok=%v n=%d err=%s", resp.OK, len(resp.Matches), resp.Error)
	}
	for i := 1; i < len(resp.Matches); i++ {
		a, b := resp.Matches[i-1], resp.Matches[i]
		if a.Round > b.Round || (a.Round == b.Round && a.ID > b.ID) {
			t.Fatalf("schedule not sorted: %s before %s", a.ID, b.ID)
		}
		if a.Played {
			t.Fatalf("match %s unexpectedly played", a.ID)
		}
	}

	// Register a round-3 match first. Later registration must not imply
	// earlier rounds were played.
	var round3ID string
	for _, m := range resp.Matches {
		if m.Round == 3 {
			round3ID = m.ID
		}
	}
	h, a := 2, 2
	_, resp = Handle(state, dir, Request{Action: "register", MatchID: round3ID, HomeGoals: &h, AwayGoals: &a})
	if !resp.OK {
		t.Fatalf("register round3: %s", resp.Error)
	}

	through := 2
	state, resp = Handle(state, dir, Request{Action: "standings", ThroughRound: &through})
	if !resp.OK {
		t.Fatalf("standings r2: %s", resp.Error)
	}
	for _, row := range resp.Standings {
		if row.Played != 0 || row.Points != 0 {
			t.Fatalf("round<=2 must be empty, got %+v", row)
		}
	}

	// Out-of-range cutoff rejected.
	over := 4
	_, resp = Handle(state, dir, Request{Action: "standings", ThroughRound: &over})
	if resp.OK {
		t.Fatal("through_round beyond range should fail")
	}

	// Full standings: the round-3 draw gives both teams one point each;
	// tie broken by team id lexicographically.
	state, resp = Handle(state, dir, Request{Action: "standings"})
	if !resp.OK {
		t.Fatalf("standings: %s", resp.Error)
	}
	if len(resp.Standings) != 4 {
		t.Fatalf("standings rows = %d, want 4", len(resp.Standings))
	}
	for _, row := range resp.Standings {
		if row.Points == 1 && row.Played != 1 {
			t.Fatalf("draw participants wrong: %+v", row)
		}
	}
	if resp.Standings[0].Points != 1 || resp.Standings[2].Points != 0 {
		t.Fatalf("unexpected table: %+v", resp.Standings)
	}

	// Correct then revoke; after revoke the table is empty again.
	h, a = 3, 0
	_, resp = Handle(state, dir, Request{Action: "correct", MatchID: round3ID, HomeGoals: &h, AwayGoals: &a})
	if !resp.OK {
		t.Fatalf("correct: %s", resp.Error)
	}
	_, resp = Handle(state, dir, Request{Action: "revoke", MatchID: round3ID})
	if !resp.OK {
		t.Fatalf("revoke: %s", resp.Error)
	}
	state, resp = Handle(state, dir, Request{Action: "standings"})
	for _, row := range resp.Standings {
		if row.Played != 0 || row.Points != 0 {
			t.Fatalf("after revoke table must be empty, got %+v", row)
		}
	}

	// Simulate a process restart: reload from disk and re-init is still
	// rejected.
	reloaded, err := Load(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded.Results) != 0 {
		t.Fatalf("results after reload = %v, want empty", reloaded.Results)
	}
	_, resp = Handle(reloaded, dir, Request{Action: "init", Teams: []string{"X", "Y"}})
	if resp.OK {
		t.Fatal("init after restart should be rejected")
	}
}

func TestProtocolErrors(t *testing.T) {
	dir := t.TempDir()
	var state *State

	_, resp := Handle(state, dir, Request{Action: "schedule"})
	if resp.OK {
		t.Fatal("schedule before init should fail")
	}
	_, resp = Handle(state, dir, Request{Action: "init", Teams: []string{"only"}})
	if resp.OK {
		t.Fatal("init with one team should fail")
	}

	state, resp = Handle(state, dir, Request{Action: "init", Teams: []string{"A", "B", "C"}})
	if !resp.OK {
		t.Fatalf("init: %s", resp.Error)
	}
	bad := -1
	_, resp = Handle(state, dir, Request{Action: "register", MatchID: state.Matches[0].ID, HomeGoals: &bad, AwayGoals: &bad})
	if resp.OK {
		t.Fatal("negative goals should fail")
	}
	_, resp = Handle(state, dir, Request{Action: "register", MatchID: "nope", HomeGoals: new(int), AwayGoals: new(int)})
	if resp.OK {
		t.Fatal("unknown match should fail")
	}
	_, resp = Handle(state, dir, Request{Action: "register", MatchID: state.Matches[0].ID})
	if resp.OK {
		t.Fatal("missing scores should fail")
	}
	_, resp = Handle(state, dir, Request{Action: "bogus"})
	if resp.OK {
		t.Fatal("unknown action should fail")
	}
}
