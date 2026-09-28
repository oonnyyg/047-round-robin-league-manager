package league

import "testing"

func validateSchedule(t *testing.T, ids []string, wantRounds int) []Match {
	t.Helper()
	matches, err := BuildSchedule(ids)
	if err != nil {
		t.Fatalf("BuildSchedule(%v): %v", ids, err)
	}
	n := len(ids)
	wantMatches := n * (n - 1) / 2
	if len(matches) != wantMatches {
		t.Fatalf("got %d matches, want %d", len(matches), wantMatches)
	}
	pairs := map[string]bool{}
	perRound := map[int]map[string]bool{}
	maxRound := 0
	for _, m := range matches {
		key := pairKey(m)
		if pairs[key] {
			t.Fatalf("pair %s plays twice", key)
		}
		pairs[key] = true
		if perRound[m.Round] == nil {
			perRound[m.Round] = map[string]bool{}
		}
		if perRound[m.Round][m.Home] || perRound[m.Round][m.Away] {
			t.Fatalf("team plays twice in round %d", m.Round)
		}
		perRound[m.Round][m.Home] = true
		perRound[m.Round][m.Away] = true
		if m.Round > maxRound {
			maxRound = m.Round
		}
	}
	if maxRound != wantRounds {
		t.Fatalf("got %d rounds, want %d", maxRound, wantRounds)
	}
	for _, a := range ids {
		for _, b := range ids {
			if a < b && !pairs[a+"\x00"+b] {
				t.Fatalf("missing pair %s vs %s", a, b)
			}
		}
	}
	return matches
}

func TestScheduleEven(t *testing.T) {
	ids := []string{"A", "B", "C", "D"}
	m1 := validateSchedule(t, ids, 3)
	m2 := validateSchedule(t, []string{"D", "A", "C", "B"}, 3)
	for i := range m1 {
		if m1[i] != m2[i] {
			t.Fatalf("schedule depends on input order:\n%v\n%v", m1, m2)
		}
	}
}

func TestScheduleOdd(t *testing.T) {
	validateSchedule(t, []string{"A", "B", "C", "D", "E"}, 5)
	validateSchedule(t, []string{"X", "Y"}, 1)
}

func TestScheduleInvalid(t *testing.T) {
	if _, err := BuildSchedule([]string{"A"}); err == nil {
		t.Fatal("want error for one team")
	}
	if _, err := BuildSchedule([]string{"A", ""}); err == nil {
		t.Fatal("want error for empty id")
	}
	if _, err := BuildSchedule([]string{"A", "A"}); err == nil {
		t.Fatal("want error for duplicate id")
	}
}
