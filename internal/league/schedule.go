package league

import (
	"sort"
	"strconv"
)

// Match is a scheduled fixture between two teams.
type Match struct {
	ID    string `json:"id"`
	Round int    `json:"round"`
	Home  string `json:"home"`
	Away  string `json:"away"`
}

// BuildSchedule creates a single round-robin schedule.
//
// Every pair of teams meets exactly once and each team plays at most once
// per round. An even number of teams yields n-1 rounds; an odd number gets a
// dummy bye slot appended, yielding n rounds with one team idle each round.
//
// The schedule depends only on the sorted team-id list, so the same teams in
// any input order always produce the same fixtures and match IDs.
func BuildSchedule(teamIDs []string) []Match {
	sorted := append([]string(nil), teamIDs...)
	sort.Strings(sorted)

	slots := append([]string(nil), sorted...)
	if len(slots)%2 == 1 {
		slots = append(slots, "") // dummy slot: the team paired with it has a bye
	}

	n := len(slots)

	type fixture struct {
		round int
		x, y  string
	}
	var planned []fixture

	totalRounds := n - 1
	for r := 0; r < totalRounds; r++ {
		var round []fixture
		for i := 0; i < n/2; i++ {
			x, y := slots[i], slots[n-1-i]
			if x == "" || y == "" {
				continue // bye, no fictitious match
			}
			round = append(round, fixture{round: r + 1, x: x, y: y})
		}
		sort.Slice(round, func(i, j int) bool {
			if round[i].x != round[j].x {
				return round[i].x < round[j].x
			}
			return round[i].y < round[j].y
		})
		planned = append(planned, round...)

		// Fix the last slot; rotate the other positions so every pair meets once.
		next := make([]string, n)
		next[0] = slots[n-2]
		for i := 1; i < n-1; i++ {
			next[i] = slots[i-1]
		}
		next[n-1] = slots[n-1]
		slots = next
	}

	matches := make([]Match, len(planned))
	width := len(strconv.Itoa(len(planned)))
	for i, f := range planned {
		matches[i] = Match{
			ID:    "M" + pad(i+1, width),
			Round: f.round,
			Home:  f.x,
			Away:  f.y,
		}
	}
	return matches
}

func pad(value, width int) string {
	s := strconv.Itoa(value)
	for len(s) < width {
		s = "0" + s
	}
	return s
}
