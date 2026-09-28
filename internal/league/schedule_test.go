package league

import (
	"reflect"
	"testing"
)

func validateSchedule(t *testing.T, teams []string, expectRounds int) []Match {
	t.Helper()
	matches := BuildSchedule(teams)

	// Exactly n*(n-1)/2 matches, unique IDs.
	wantCount := len(teams) * (len(teams) - 1) / 2
	if len(matches) != wantCount {
		t.Fatalf("match count = %d, want %d", len(matches), wantCount)
	}
	ids := map[string]bool{}
	pairs := map[string]bool{}
	roundTeams := map[int]map[string]bool{}
	maxRound := 0
	for _, m := range matches {
		if ids[m.ID] {
			t.Fatalf("duplicate match id %q", m.ID)
		}
		ids[m.ID] = true
		key := pairKey(m.Home, m.Away)
		if pairs[key] {
			t.Fatalf("pair %s meets twice", key)
		}
		pairs[key] = true
		if roundTeams[m.Round] == nil {
			roundTeams[m.Round] = map[string]bool{}
		}
		if roundTeams[m.Round][m.Home] || roundTeams[m.Round][m.Away] {
			t.Fatalf("team plays twice in round %d", m.Round)
		}
		roundTeams[m.Round][m.Home] = true
		roundTeams[m.Round][m.Away] = true
		if m.Round > maxRound {
			maxRound = m.Round
		}
	}
	if maxRound != expectRounds {
		t.Fatalf("rounds = %d, want %d", maxRound, expectRounds)
	}

	// IDs sorted by round then id; ids in round order.
	for i := 1; i < len(matches); i++ {
		if matches[i].Round < matches[i-1].Round {
			t.Fatalf("matches not sorted by round at %d", i)
		}
		if matches[i].Round == matches[i-1].Round && matches[i].ID < matches[i-1].ID {
			t.Fatalf("matches not sorted by id in round at %d", i)
		}
	}
	return matches
}

func pairKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func TestBuildScheduleEven(t *testing.T) {
	teams := []string{"A", "B", "C", "D"}
	matches := validateSchedule(t, teams, 3)
	// Each round has n/2 matches.
	counts := map[int]int{}
	for _, m := range matches {
		counts[m.Round]++
	}
	for r := 1; r <= 3; r++ {
		if counts[r] != 2 {
			t.Fatalf("round %d has %d matches, want 2", r, counts[r])
		}
	}
}

func TestBuildScheduleOdd(t *testing.T) {
	teams := []string{"A", "B", "C", "D", "E"}
	matches := validateSchedule(t, teams, 5)
	// Each round has (n-1)/2 real matches; one team is idle.
	counts := map[int]int{}
	for _, m := range matches {
		counts[m.Round]++
	}
	for r := 1; r <= 5; r++ {
		if counts[r] != 2 {
			t.Fatalf("round %d has %d matches, want 2", r, counts[r])
		}
	}
}

func TestBuildScheduleOrderIndependent(t *testing.T) {
	a := BuildSchedule([]string{"A", "B", "C", "D"})
	b := BuildSchedule([]string{"D", "A", "C", "B"})
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("schedules differ by input order:\n%v\n%v", a, b)
	}
}

func TestBuildScheduleStableIDs(t *testing.T) {
	matches := BuildSchedule([]string{"t1", "t2", "t3", "t4", "t5", "t6"})
	if matches[0].ID != "M01" || matches[len(matches)-1].ID != "M15" {
		t.Fatalf("unexpected ids: %s .. %s", matches[0].ID, matches[len(matches)-1].ID)
	}
}
