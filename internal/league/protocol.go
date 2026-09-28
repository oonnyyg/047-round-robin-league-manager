package league

import (
	"encoding/json"
)

// Request is one JSON request line. The action field selects the operation.
type Request struct {
	Action       string   `json:"action"`
	Teams        []string `json:"teams,omitempty"`
	MatchID      string   `json:"match_id,omitempty"`
	HomeGoals    *int     `json:"home_goals,omitempty"`
	AwayGoals    *int     `json:"away_goals,omitempty"`
	ThroughRound *int     `json:"through_round,omitempty"`
}

// Response is one JSON response line.
type Response struct {
	OK          bool            `json:"ok"`
	Error       string          `json:"error,omitempty"`
	Teams       []string        `json:"teams,omitempty"`
	TotalRounds int             `json:"total_rounds,omitempty"`
	Matches     []ScheduleEntry `json:"matches,omitempty"`
	Standings   []Row           `json:"standings,omitempty"`
}

// ScheduleEntry is a fixture plus any registered score.
type ScheduleEntry struct {
	ID        string `json:"id"`
	Round     int    `json:"round"`
	Home      string `json:"home"`
	Away      string `json:"away"`
	Played    bool   `json:"played"`
	HomeGoals *int   `json:"home_goals,omitempty"`
	AwayGoals *int   `json:"away_goals,omitempty"`
}

// Handle processes a single request against the loaded state, persisting
// mutating operations to dir. It returns an OK=false response for expected
// failures without changing any state.
func Handle(s *State, dir string, req Request) (*State, Response) {
	switch req.Action {
	case "init":
		if s != nil {
			return s, fail("league already initialized; re-init is not allowed")
		}
		next, err := New(req.Teams)
		if err != nil {
			return nil, fail(err.Error())
		}
		if err := next.Save(dir); err != nil {
			return nil, fail(err.Error())
		}
		return next, Response{OK: true, Teams: next.Teams, TotalRounds: next.TotalRounds()}

	case "schedule":
		if s == nil {
			return nil, fail("league not initialized")
		}
		entries := make([]ScheduleEntry, 0, len(s.Matches))
		for _, m := range s.Matches {
			e := ScheduleEntry{ID: m.ID, Round: m.Round, Home: m.Home, Away: m.Away}
			if r, ok := s.Results[m.ID]; ok {
				h, a := r.HomeGoals, r.AwayGoals
				e.Played = true
				e.HomeGoals = &h
				e.AwayGoals = &a
			}
			entries = append(entries, e)
		}
		return s, Response{OK: true, Teams: s.Teams, TotalRounds: s.TotalRounds(), Matches: entries}

	case "register", "correct":
		if s == nil {
			return nil, fail("league not initialized")
		}
		if req.HomeGoals == nil || req.AwayGoals == nil {
			return s, fail("home_goals and away_goals are required non-negative integers")
		}
		var err error
		if req.Action == "register" {
			err = s.Register(req.MatchID, *req.HomeGoals, *req.AwayGoals)
		} else {
			err = s.Correct(req.MatchID, *req.HomeGoals, *req.AwayGoals)
		}
		if err != nil {
			return s, fail(err.Error())
		}
		if err := s.Save(dir); err != nil {
			return s, fail(err.Error())
		}
		return s, Response{OK: true}

	case "revoke":
		if s == nil {
			return nil, fail("league not initialized")
		}
		if err := s.Revoke(req.MatchID); err != nil {
			return s, fail(err.Error())
		}
		if err := s.Save(dir); err != nil {
			return s, fail(err.Error())
		}
		return s, Response{OK: true}

	case "standings":
		if s == nil {
			return nil, fail("league not initialized")
		}
		rows, err := s.Standings(req.ThroughRound)
		if err != nil {
			return s, fail(err.Error())
		}
		return s, Response{OK: true, TotalRounds: s.TotalRounds(), Standings: rows}

	default:
		return s, fail("unknown action: " + req.Action)
	}
}

func fail(msg string) Response {
	return Response{OK: false, Error: msg}
}

// EncodeResponse serializes a response as a single JSON line.
func EncodeResponse(resp Response) ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
