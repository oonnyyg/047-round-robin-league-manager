package league

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const stateFileName = "league-state.json"

// StatePath returns the state file path inside the given data directory.
func StatePath(dir string) string {
	return filepath.Join(dir, stateFileName)
}

// Exists reports whether a league state file is present in dir.
func Exists(dir string) bool {
	_, err := os.Stat(StatePath(dir))
	return err == nil
}

// Save writes the league to dir/league-state.json. A temporary file is
// renamed into place so a crash cannot leave a truncated state file.
func (l *League) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	final := StatePath(dir)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// Load reads a league previously saved in dir.
func Load(dir string) (*League, error) {
	data, err := os.ReadFile(StatePath(dir))
	if err != nil {
		return nil, err
	}
	var l League
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	if l.Results == nil {
		l.Results = map[string]Result{}
	}
	if len(l.Teams) < 2 || len(l.Matches) == 0 {
		return nil, errors.New("state file is not a valid league")
	}
	return &l, nil
}

// ErrNotExist is returned by Load when no state file exists.
var ErrNotExist = fs.ErrNotExist
