package league

import (
	"fmt"
	"sort"
)

// Result is a registered score for a match.
type Result struct {
	HomeGoals int `json:"home_goals"`
	AwayGoals int `json:"away_goals"`
}

// State is the complete persisted state of a league.
type State struct {
	Teams   []string          `json:"teams"`
	Matches []Match           `json:"matches"`
	Results map[string]Result `json:"results"`
}

// New creates a league with a fixed team list and schedule.
func New(teamIDs []string) (*State, error) {
	matches, err := BuildSchedule(teamIDs)
	if err != nil {
		return nil, err
	}
	ids := append([]string(nil), teamIDs...)
	sort.Strings(ids)
	return &State{
		Teams:   ids,
		Matches: matches,
		Results: map[string]Result{},
	}, nil
}

// TotalRounds returns the number of rounds in the schedule.
func (s *State) TotalRounds() int {
	max := 0
	for _, m := range s.Matches {
		if m.Round > max {
			max = m.Round
		}
	}
	return max
}

func (s *State) match(id string) (int, bool) {
	for i := range s.Matches {
		if s.Matches[i].ID == id {
			return i, true
		}
	}
	return 0, false
}

func validateScore(home, away int) error {
	if home < 0 || away < 0 {
		return fmt.Errorf("goals must be non-negative integers")
	}
	return nil
}

// Register records a result for a match that has none yet.
func (s *State) Register(matchID string, home, away int) error {
	if err := validateScore(home, away); err != nil {
		return err
	}
	if _, ok := s.match(matchID); !ok {
		return fmt.Errorf("unknown match: %s", matchID)
	}
	if _, ok := s.Results[matchID]; ok {
		return fmt.Errorf("match %s already has a result", matchID)
	}
	s.Results[matchID] = Result{HomeGoals: home, AwayGoals: away}
	return nil
}

// Correct replaces the score of a match that already has a result.
func (s *State) Correct(matchID string, home, away int) error {
	if err := validateScore(home, away); err != nil {
		return err
	}
	if _, ok := s.match(matchID); !ok {
		return fmt.Errorf("unknown match: %s", matchID)
	}
	if _, ok := s.Results[matchID]; !ok {
		return fmt.Errorf("match %s has no result to correct", matchID)
	}
	s.Results[matchID] = Result{HomeGoals: home, AwayGoals: away}
	return nil
}

// Revoke removes the result of a match, returning it to unplayed status.
func (s *State) Revoke(matchID string) error {
	if _, ok := s.match(matchID); !ok {
		return fmt.Errorf("unknown match: %s", matchID)
	}
	if _, ok := s.Results[matchID]; !ok {
		return fmt.Errorf("match %s has no result to revoke", matchID)
	}
	delete(s.Results, matchID)
	return nil
}
