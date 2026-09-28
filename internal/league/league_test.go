package league

import "testing"

func mustNew(t *testing.T, ids []string) *State {
	t.Helper()
	s, err := New(ids)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestRegisterCorrectRevoke(t *testing.T) {
	s := mustNew(t, []string{"A", "B", "C", "D"})
	mid := s.Matches[0].ID

	if err := s.Register(mid, 1, 0); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := s.Register(mid, 2, 2); err == nil {
		t.Fatal("duplicate register should fail")
	}
	if err := s.Register("M99-1", 1, 1); err == nil {
		t.Fatal("unknown match should fail")
	}
	if err := s.Register(s.Matches[1].ID, -1, 0); err == nil {
		t.Fatal("negative score should fail")
	}
	if got := s.Results[mid]; got.HomeGoals != 1 || got.AwayGoals != 0 {
		t.Fatalf("failed register changed state: %+v", got)
	}

	if err := s.Correct(mid, 2, 2); err != nil {
		t.Fatalf("correct: %v", err)
	}
	if got := s.Results[mid]; got.HomeGoals != 2 || got.AwayGoals != 2 {
		t.Fatalf("correct not applied: %+v", got)
	}
	if err := s.Correct(s.Matches[1].ID, 1, 0); err == nil {
		t.Fatal("correct unplayed should fail")
	}

	if err := s.Revoke(mid); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, ok := s.Results[mid]; ok {
		t.Fatal("revoke did not remove result")
	}
	if err := s.Revoke(mid); err == nil {
		t.Fatal("revoke unplayed should fail")
	}
}

func TestStandingsScoringAndOrder(t *testing.T) {
	s := mustNew(t, []string{"A", "B", "C"})
	byTeams := map[[2]string]string{}
	for _, m := range s.Matches {
		byTeams[[2]string{m.Home, m.Away}] = m.ID
	}
	// A beats B 3-0; B draws C 1-1; A beats C 1-0.
	if err := s.Register(byTeams[[2]string{"A", "B"}], 3, 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Register(byTeams[[2]string{"B", "C"}], 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Register(byTeams[[2]string{"A", "C"}], 1, 0); err != nil {
		t.Fatal(err)
	}
	rows, err := s.Standings(nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Row{
		"A": {Team: "A", Played: 2, Wins: 2, GoalsFor: 4, GoalsAgainst: 0, GoalDiff: 4, Points: 6},
		"B": {Team: "B", Played: 2, Draws: 1, Losses: 1, GoalsFor: 1, GoalsAgainst: 4, GoalDiff: -3, Points: 1},
		"C": {Team: "C", Played: 2, Draws: 1, Losses: 1, GoalsFor: 1, GoalsAgainst: 2, GoalDiff: -1, Points: 1},
	}
	for _, r := range rows {
		w := want[r.Team]
		if r != w {
			t.Fatalf("row for %s = %+v, want %+v", r.Team, r, w)
		}
	}
	if rows[0].Team != "A" || rows[1].Team != "C" || rows[2].Team != "B" {
		t.Fatalf("order = %s,%s,%s, want A,C,B", rows[0].Team, rows[1].Team, rows[2].Team)
	}
}

func TestStandingsThroughRound(t *testing.T) {
	// Four teams, three rounds. Register only a later-round match and check
	// earlier rounds do not infer completion.
	s := mustNew(t, []string{"A", "B", "C", "D"})
	later := ""
	for _, m := range s.Matches {
		if m.Round == 3 {
			later = m.ID
			if err := s.Register(m.ID, 5, 0); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	r1, err := s.Standings(ptrInt(1))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range r1 {
		if r.Played != 0 || r.Points != 0 {
			t.Fatalf("round 1 table should be empty, got %+v", r)
		}
	}
	all, err := s.Standings(nil)
	if err != nil {
		t.Fatal(err)
	}
	played := 0
	for _, r := range all {
		played += r.Played
	}
	if played != 2 {
		t.Fatalf("full table should include one match (2 player-slots), got %d", played)
	}

	if _, err := s.Standings(ptrInt(0)); err == nil {
		t.Fatal("through_round 0 should fail")
	}
	if _, err := s.Standings(ptrInt(4)); err == nil {
		t.Fatal("through_round beyond schedule should fail")
	}
	_ = later
}

func ptrInt(i int) *int { return &i }
