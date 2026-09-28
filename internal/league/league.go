package league

import (
	"errors"
	"sort"
)

// Result is a registered score for a match.
type Result struct {
	HomeGoals int `json:"home_goals"`
	AwayGoals int `json:"away_goals"`
}

// League holds the fixed team list, schedule and registered results.
type League struct {
	Teams   []string          `json:"teams"`
	Matches []Match           `json:"matches"`
	Results map[string]Result `json:"results"`
}

// New validates the team list and initializes a league with its schedule.
func New(teamIDs []string) (*League, error) {
	if len(teamIDs) < 2 {
		return nil, errors.New("at least two teams are required")
	}
	seen := map[string]bool{}
	for _, id := range teamIDs {
		if id == "" {
			return nil, errors.New("team id must not be empty")
		}
		if seen[id] {
			return nil, errors.New("duplicate team id: " + id)
		}
		seen[id] = true
	}
	teams := append([]string(nil), teamIDs...)
	sort.Strings(teams)
	return &League{
		Teams:   teams,
		Matches: BuildSchedule(teamIDs),
		Results: map[string]Result{},
	}, nil
}

func (l *League) match(id string) (int, bool) {
	// Schedule is small; a linear scan keeps the structure simple.
	for i := range l.Matches {
		if l.Matches[i].ID == id {
			return i, true
		}
	}
	return 0, false
}

// Register records a score for a match that has no result yet.
func (l *League) Register(matchID string, homeGoals, awayGoals int) error {
	if _, ok := l.match(matchID); !ok {
		return errors.New("unknown match: " + matchID)
	}
	if homeGoals < 0 || awayGoals < 0 {
		return errors.New("goals must be non-negative integers")
	}
	if _, ok := l.Results[matchID]; ok {
		return errors.New("match already has a result, use correct instead: " + matchID)
	}
	l.Results[matchID] = Result{HomeGoals: homeGoals, AwayGoals: awayGoals}
	return nil
}

// Correct replaces the score of an already registered match.
func (l *League) Correct(matchID string, homeGoals, awayGoals int) error {
	if _, ok := l.match(matchID); !ok {
		return errors.New("unknown match: " + matchID)
	}
	if homeGoals < 0 || awayGoals < 0 {
		return errors.New("goals must be non-negative integers")
	}
	if _, ok := l.Results[matchID]; !ok {
		return errors.New("match has no result yet, use register instead: " + matchID)
	}
	l.Results[matchID] = Result{HomeGoals: homeGoals, AwayGoals: awayGoals}
	return nil
}

// Revoke removes the result of a registered match, restoring unplayed state.
func (l *League) Revoke(matchID string) error {
	if _, ok := l.match(matchID); !ok {
		return errors.New("unknown match: " + matchID)
	}
	if _, ok := l.Results[matchID]; !ok {
		return errors.New("match has no result to revoke: " + matchID)
	}
	delete(l.Results, matchID)
	return nil
}

// MaxRound returns the number of rounds in the schedule.
func (l *League) MaxRound() int {
	max := 0
	for _, m := range l.Matches {
		if m.Round > max {
			max = m.Round
		}
	}
	return max
}

// Row is one team's standing line.
type Row struct {
	Team           string `json:"team"`
	Played         int    `json:"played"`
	Wins           int    `json:"wins"`
	Draws          int    `json:"draws"`
	Losses         int    `json:"losses"`
	GoalsFor       int    `json:"goals_for"`
	GoalsAgainst   int    `json:"goals_against"`
	GoalDifference int    `json:"goal_difference"`
	Points         int    `json:"points"`
}

// Standings computes the table. throughRound <= 0 means all rounds.
// A positive throughRound beyond the schedule is rejected.
func (l *League) Standings(throughRound int) ([]Row, error) {
	maxRound := l.MaxRound()
	if throughRound < 0 {
		return nil, errors.New("through round must be a positive integer")
	}
	if throughRound > maxRound {
		return nil, errors.New("through round exceeds schedule range")
	}

	rows := map[string]*Row{}
	for _, team := range l.Teams {
		rows[team] = &Row{Team: team}
	}

	for _, m := range l.Matches {
		if throughRound > 0 && m.Round > throughRound {
			continue
		}
		res, ok := l.Results[m.ID]
		if !ok {
			continue // unplayed matches count for nothing
		}
		home := rows[m.Home]
		away := rows[m.Away]
		home.Played++
		away.Played++
		home.GoalsFor += res.HomeGoals
		home.GoalsAgainst += res.AwayGoals
		away.GoalsFor += res.AwayGoals
		away.GoalsAgainst += res.HomeGoals
		switch {
		case res.HomeGoals > res.AwayGoals:
			home.Wins++
			home.Points += 3
			away.Losses++
		case res.HomeGoals < res.AwayGoals:
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

	out := make([]Row, 0, len(l.Teams))
	for _, r := range rows {
		r.GoalDifference = r.GoalsFor - r.GoalsAgainst
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Points != out[j].Points {
			return out[i].Points > out[j].Points
		}
		if out[i].GoalDifference != out[j].GoalDifference {
			return out[i].GoalDifference > out[j].GoalDifference
		}
		if out[i].GoalsFor != out[j].GoalsFor {
			return out[i].GoalsFor > out[j].GoalsFor
		}
		return out[i].Team < out[j].Team
	})
	return out, nil
}
