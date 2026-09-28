package league

import (
	"fmt"
	"sort"
)

// Row is one team's line in the standings table.
type Row struct {
	Team         string `json:"team"`
	Played       int    `json:"played"`
	Wins         int    `json:"wins"`
	Draws        int    `json:"draws"`
	Losses       int    `json:"losses"`
	GoalsFor     int    `json:"goals_for"`
	GoalsAgainst int    `json:"goals_against"`
	GoalDiff     int    `json:"goal_diff"`
	Points       int    `json:"points"`
}

// Standings computes the table from registered results. If throughRound is
// nil all rounds count; otherwise only rounds <= throughRound count. A
// throughRound beyond the schedule range is rejected.
func (s *State) Standings(throughRound *int) ([]Row, error) {
	limit := s.TotalRounds()
	if throughRound != nil {
		if *throughRound <= 0 {
			return nil, fmt.Errorf("through_round must be a positive integer")
		}
		if *throughRound > limit {
			return nil, fmt.Errorf("through_round %d exceeds total rounds %d", *throughRound, limit)
		}
		limit = *throughRound
	}

	rows := make(map[string]*Row, len(s.Teams))
	for _, id := range s.Teams {
		rows[id] = &Row{Team: id}
	}
	for _, m := range s.Matches {
		if m.Round > limit {
			continue
		}
		r, ok := s.Results[m.ID]
		if !ok {
			continue
		}
		home := rows[m.Home]
		away := rows[m.Away]
		home.Played++
		away.Played++
		home.GoalsFor += r.HomeGoals
		home.GoalsAgainst += r.AwayGoals
		away.GoalsFor += r.AwayGoals
		away.GoalsAgainst += r.HomeGoals
		switch {
		case r.HomeGoals > r.AwayGoals:
			home.Wins++
			home.Points += 3
			away.Losses++
		case r.HomeGoals < r.AwayGoals:
			away.Wins++
			away.Points += 3
			home.Losses++
		default:
			home.Draws++
			away.Draws++
			home.Points++
			away.Points++
		}
	}

	out := make([]Row, 0, len(s.Teams))
	for _, r := range rows {
		r.GoalDiff = r.GoalsFor - r.GoalsAgainst
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Points != out[j].Points {
			return out[i].Points > out[j].Points
		}
		if out[i].GoalDiff != out[j].GoalDiff {
			return out[i].GoalDiff > out[j].GoalDiff
		}
		if out[i].GoalsFor != out[j].GoalsFor {
			return out[i].GoalsFor > out[j].GoalsFor
		}
		return out[i].Team < out[j].Team
	})
	return out, nil
}
