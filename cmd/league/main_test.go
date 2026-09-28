package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func feed(t *testing.T, dir, input string) []response {
	t.Helper()
	var in, out bytes.Buffer
	in.WriteString(input)
	if err := run(dir, &in, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got []response
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var r response
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			t.Fatalf("bad response %q: %v", line, err)
		}
		got = append(got, r)
	}
	return got
}

func TestProtocolSessionAndReopen(t *testing.T) {
	dir := t.TempDir()

	first := feed(t, dir, strings.Join([]string{
		`{"op":"init","teams":["D","A","C","B"]}`,
		`{"op":"init","teams":["A","B"]}`,
		`{"op":"schedule"}`,
		`{"op":"register","match_id":"M1","home_goals":1,"away_goals":0}`,
		`{"op":"register","match_id":"M1","home_goals":9,"away_goals":9}`,
		`{"op":"register","match_id":"ZZZ","home_goals":1,"away_goals":0}`,
		`{"op":"register","match_id":"M2","home_goals":-1,"away_goals":0}`,
		`{"op":"correct","match_id":"M1","home_goals":2,"away_goals":2}`,
		`{"op":"standings"}`,
		`{"op":"standings","through_round":1}`,
		`{"op":"standings","through_round":99}`,
		`{"op":"revoke","match_id":"M1"}`,
		`{"op":"revoke","match_id":"M1"}`,
		`{"op":"bogus"}`,
		`{bad json`,
	}, "\n")+"\n")

	if !first[0].OK {
		t.Fatalf("init failed: %s", first[0].Error)
	}
	if first[1].OK {
		t.Fatal("duplicate init accepted")
	}
	if len(first[2].Matches) != 6 {
		t.Fatalf("schedule matches = %d, want 6", len(first[2].Matches))
	}
	if !first[3].OK {
		t.Fatalf("register failed: %s", first[3].Error)
	}
	if first[4].OK || first[5].OK || first[6].OK {
		t.Fatal("invalid register accepted")
	}
	if !first[7].OK {
		t.Fatalf("correct failed: %s", first[7].Error)
	}
	// After correction M1 is a 2-2 draw.
	if first[8].Standings[0].Points != 1 || first[8].Standings[0].Draws != 1 {
		t.Fatalf("unexpected standings row: %+v", first[8].Standings[0])
	}
	// Round 1 contains M1 only (schedule is ordered by round then id).
	if first[9].Standings[0].Played != 1 {
		t.Fatal("through_round=1 did not include round 1 result")
	}
	if first[10].OK {
		t.Fatal("out-of-range through_round accepted")
	}
	if !first[11].OK {
		t.Fatalf("revoke failed: %s", first[11].Error)
	}
	if first[12].OK {
		t.Fatal("revoke of unplayed match accepted")
	}
	if first[13].OK || first[14].OK {
		t.Fatal("bad op / bad json should fail")
	}

	// Reopen: revoked result must stay gone, team list still present.
	reopened := feed(t, dir, `{"op":"schedule"}`+"\n")
	if !reopened[0].OK || len(reopened[0].Matches) != 6 {
		t.Fatalf("reopen schedule unexpected: %+v", reopened[0])
	}
	for _, m := range reopened[0].Matches {
		if m.Played {
			t.Fatalf("match %s still played after reopen", m.ID)
		}
	}
}
