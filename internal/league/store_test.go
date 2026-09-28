package league

import (
	"errors"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Load before init = %v, want ErrNotInitialized", err)
	}

	s := mustNew(t, []string{"A", "B", "C", "D"})
	if err := s.Register(s.Matches[0].ID, 2, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got.Teams) != len(s.Teams) || len(got.Matches) != len(s.Matches) {
		t.Fatal("reloaded teams/matches differ")
	}
	if r, ok := got.Results[s.Matches[0].ID]; !ok || r.HomeGoals != 2 || r.AwayGoals != 1 {
		t.Fatalf("reloaded results wrong: %+v", got.Results)
	}
	rows, err := got.Standings(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Points != 3 {
		t.Fatalf("standings after reload wrong: %+v", rows[0])
	}
}
