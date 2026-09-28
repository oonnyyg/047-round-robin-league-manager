package league

import (
	"fmt"
	"sort"
)

// Match is a single scheduled fixture between two teams.
type Match struct {
	ID    string `json:"id"`
	Round int    `json:"round"`
	Home  string `json:"home"`
	Away  string `json:"away"`
}

// BuildSchedule builds a single round-robin schedule using the circle method.
// Every pair of teams meets exactly once and each team plays at most once per
// round. Teams are sorted by ID first, so the same set of teams always yields
// the same schedule regardless of input order. With an odd number of teams a
// bye team is skipped each round (no match is created for it).
func BuildSchedule(teamIDs []string) ([]Match, error) {
	if len(teamIDs) < 2 {
		return nil, fmt.Errorf("at least two teams are required")
	}
	seen := make(map[string]struct{}, len(teamIDs))
	ids := make([]string, 0, len(teamIDs))
	for _, id := range teamIDs {
		if id == "" {
			return nil, fmt.Errorf("team id must not be empty")
		}
		if _, dup := seen[id]; dup {
			return nil, fmt.Errorf("duplicate team id: %s", id)
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Circle method: keep the first element fixed and rotate the rest. For
	// an odd number of teams add a bye marker; rounds == number of teams.
	n := len(ids)
	circle := make([]string, n)
	copy(circle, ids)
	rounds := n - 1
	if n%2 == 1 {
		circle = append(circle, "") // bye marker
		rounds = n
	}

	width := len(circle)
	matches := make([]Match, 0, n*(n-1)/2)
	for round := 1; round <= rounds; round++ {
		var roundMatches []Match
		for i := 0; i < width/2; i++ {
			home := circle[i]
			away := circle[width-1-i]
			if home == "" || away == "" {
				continue // one team has a bye this round
			}
			roundMatches = append(roundMatches, Match{
				Round: round,
				Home:  home,
				Away:  away,
			})
		}
		// Deterministic order within a round: lexicographic pair key.
		sort.Slice(roundMatches, func(i, j int) bool {
			return pairKey(roundMatches[i]) < pairKey(roundMatches[j])
		})
		for i := range roundMatches {
			roundMatches[i].ID = fmt.Sprintf("M%02d-%d", round, i+1)
		}
		matches = append(matches, roundMatches...)

		// Rotate: keep circle[0] fixed, shift the others clockwise.
		last := circle[width-1]
		copy(circle[2:], circle[1:width-1])
		circle[1] = last
	}
	return matches, nil
}

func pairKey(m Match) string {
	if m.Home < m.Away {
		return m.Home + "\x00" + m.Away
	}
	return m.Away + "\x00" + m.Home
}
