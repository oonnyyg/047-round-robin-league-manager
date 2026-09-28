# Round-Robin League Manager

A fully local league scheduler and standings tracker. It generates a
single round-robin schedule (every pair of teams meets exactly once) and keeps
match results and standings consistent as results are registered, corrected
and revoked. Only the Go standard library is used; no network access is
required.

## Run

```sh
go run ./cmd/league --state ./data
```

The program reads one JSON request per line from standard input and writes one
JSON response per line to standard output. `--state` names the data directory
where `league-state.json` is stored (created on first initialization). Exit at
end of input (Ctrl-Z / Ctrl-D or close the pipe); reopen with the same
`--state` directory to continue the league.

Try the bundled example:

```sh
go run ./cmd/league --state ./data < examples/session.jsonl
```

## Requests

All requests are JSON objects with an `op` field.

| op          | extra fields                                             | meaning |
|-------------|----------------------------------------------------------|---------|
| `init`      | `teams`: array of non-empty, unique team ids (>= 2)      | create the league and schedule |
| `schedule`  | none                                                     | list teams and all matches |
| `register`  | `match_id`, `home_goals`, `away_goals`                   | add a result to an unplayed match |
| `correct`   | `match_id`, `home_goals`, `away_goals`                   | replace an existing result |
| `revoke`    | `match_id`                                               | delete a result (match becomes unplayed) |
| `standings` | optional `through_round`: positive integer               | table for all rounds, or rounds <= cutoff |

Goals must be non-negative integers. Every response is
`{"ok": true, ...}` or `{"ok": false, "error": "..."}`. A failed request
never changes stored results.

Examples:

```json
{"op":"init","teams":["A","B","C"]}
{"op":"register","match_id":"M2","home_goals":2,"away_goals":1}
{"op":"correct","match_id":"M2","home_goals":0,"away_goals":0}
{"op":"revoke","match_id":"M2"}
{"op":"standings","through_round":2}
```

## Rules

- **Initialization**: at least two teams with non-empty, unique ids. A league
  can be initialized only once per data directory; a second `init` is
  rejected. Teams and fixtures never change afterward.
- **Schedule**: even number of teams -> n-1 rounds; odd number -> n rounds
  with one team idle (bye) each round. Byes produce no match. Each team plays
  at most once per round.
- **Stable IDs**: matches are numbered `M1`, `M2`, ... (zero-padded only when
  there are ten or more matches, e.g. `M01`). The schedule is derived from the
  sorted team-id list, so identical rosters in any input order produce the
  identical schedule and IDs.
- **Ordering**: the schedule is sorted by round, then by match id.
- **Results**: `register` only accepts unplayed matches; `correct` and
  `revoke` only accept played matches. Unknown match ids, negative scores and
  state mismatches are rejected.
- **Standings**: only played matches count. Win = 3 points, draw = 1 each,
  loss = 0. Rows include `played`, `wins`, `draws`, `losses`, `goals_for`,
  `goals_against`, `goal_difference` and `points`.
- **Standings order**: points desc, goal difference desc, goals for desc,
  then team id lexicographic asc.
- **Cutoff rounds**: `through_round` includes only played matches in that
  round or earlier; registering a later round does not imply earlier matches
  were played. A cutoff larger than the number of scheduled rounds (or
  non-positive) is rejected, never silently truncated.
- **Persistence**: every successful mutation is saved immediately via a
  temp-file-and-rename write, so a crash cannot corrupt the state file.

## Layout

- `cmd/league` — line-based JSON entry point.
- `internal/league` — scheduler, league state, result operations, standings
  and persistence.
- `examples/session.jsonl` — sample request stream.

## Test

```sh
go test ./...
```

Tests cover even/odd schedules (pair uniqueness, one game per round, byes),
input-order independence, draw ties and tie-breaks, register/correct/revoke,
round cutoffs, protocol errors and reopening a saved league.
