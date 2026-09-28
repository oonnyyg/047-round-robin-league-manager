package league

import "testing"

func mustNewLeague(t *testing.T, teams []string) *League {
	t.Helper()
	l, err := New(teams)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return l
}

func TestNewValidation(t *testing.T) {
	if _, err := New([]string{"a"}); err == nil {
		t.Fatal("expected error for one team")
	}
	if _, err := New([]string{"a", ""}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := New([]string{"a", "a"}); err == nil {
		t.Fatal("expected error for duplicate id")
	}
}

func TestRegisterCorrectRevoke(t *testing.T) {
	l := mustNewLeague(t, []string{"A", "B", "C"})
	m := l.Matches[0].ID

	if err := l.Register("nope", 1, 0); err == nil {
		t.Fatal("unknown match accepted")
	}
	if err := l.Register(m, -1, 0); err == nil {
		t.Fatal("negative score accepted")
	}
	if err := l.Register(m, 1, 0); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := l.Register(m, 2, 0); err == nil {
		t.Fatal("duplicate register accepted")
	}
	if err := l.Correct(m, 2, 2); err != nil {
		t.Fatalf("correct: %v", err)
	}
	if got := l.Results[m]; got != (Result{2, 2}) {
		t.Fatalf("result = %v, want {2 2}", got)
	}
	if err := l.Correct("nope", 0, 0); err == nil {
		t.Fatal("correct unknown match accepted")
	}
	if err := l.Revoke(m); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, ok := l.Results[m]; ok {
		t.Fatal("revoke did not remove result")
	}
	if err := l.Revoke(m); err == nil {
		t.Fatal("revoke unplayed match accepted")
	}
}

func TestStandingsScoringAndOrder(t *testing.T) {
	l := mustNewLeague(t, []string{"A", "B", "C", "D"})
	byPair := map[string]string{}
	for _, m := range l.Matches {
		byPair[pairKey(m.Home, m.Away)] = m.ID
	}
	// A beats B, C draws D, A draws C, B beats D, A beats D, B draws C.
	l.Register(byPair["A|B"], 1, 0)
	l.Register(byPair["C|D"], 2, 2)
	l.Register(byPair["A|C"], 0, 0)
	l.Register(byPair["B|D"], 3, 1)
	l.Register(byPair["A|D"], 2, 0)
	l.Register(byPair["B|C"], 1, 1)

	rows, err := l.Standings(0)
	if err != nil {
		t.Fatalf("standings: %v", err)
	}
	want := map[string]Row{}
	want["A"] = Row{Team: "A", Played: 3, Wins: 2, Draws: 1, GoalsFor: 3, GoalsAgainst: 0, GoalDifference: 3, Points: 7}
	want["B"] = Row{Team: "B", Played: 3, Wins: 1, Draws: 1, Losses: 1, GoalsFor: 4, GoalsAgainst: 3, GoalDifference: 1, Points: 4}
	want["C"] = Row{Team: "C", Played: 3, Wins: 0, Draws: 3, GoalsFor: 3, GoalsAgainst: 3, GoalDifference: 0, Points: 3}
	want["D"] = Row{Team: "D", Played: 3, Wins: 0, Draws: 1, Losses: 2, GoalsFor: 3, GoalsAgainst: 7, GoalDifference: -4, Points: 1}
	for _, got := range rows {
		w := want[got.Team]
		if got != w {
			t.Fatalf("row %s = %+v, want %+v", got.Team, got, w)
		}
	}
	if rows[0].Team != "A" || rows[1].Team != "B" || rows[2].Team != "C" || rows[3].Team != "D" {
		t.Fatalf("order = %v", teamsOf(rows))
	}
}

func TestStandingsTieBreakByTeamID(t *testing.T) {
	l := mustNewLeague(t, []string{"X", "Y"})
	l.Register(l.Matches[0].ID, 0, 0)
	rows, _ := l.Standings(0)
	if rows[0].Team != "X" || rows[1].Team != "Y" {
		t.Fatalf("tie order = %v, want X then Y", teamsOf(rows))
	}
	for _, r := range rows {
		if r.Points != 1 || r.Played != 1 || r.Draws != 1 {
			t.Fatalf("unexpected tied row %+v", r)
		}
	}
}

func TestStandingsThroughRound(t *testing.T) {
	l := mustNewLeague(t, []string{"A", "B", "C", "D"})
	// Register only round 1 and 3 matches; round 2 stays unplayed.
	var r1, r3 string
	for _, m := range l.Matches {
		switch m.Round {
		case 1:
			r1 = m.ID
		case 3:
			r3 = m.ID
		}
	}
	l.Register(r1, 1, 0)
	l.Register(r3, 5, 0)

	rows, err := l.Standings(1)
	if err != nil {
		t.Fatalf("standings: %v", err)
	}
	played := 0
	for _, r := range rows {
		played += r.Played
	}
	if played != 2 {
		t.Fatalf("played sum = %d, want 2 (round 1 only)", played)
	}

	if _, err := l.Standings(l.MaxRound() + 1); err == nil {
		t.Fatal("out-of-range through round accepted")
	}
	if _, err := l.Standings(-2); err == nil {
		t.Fatal("negative through round accepted")
	}
}

func teamsOf(rows []Row) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Team
	}
	return out
}
