package league

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const stateFileName = "league.json"

// ErrNotInitialized is returned by Load when the data directory has no league.
var ErrNotInitialized = errors.New("league not initialized")

// Load reads the league state from dir. It returns ErrNotInitialized if the
// state file does not exist yet.
func Load(dir string) (*State, error) {
	path := filepath.Join(dir, stateFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	if s.Results == nil {
		s.Results = map[string]Result{}
	}
	return &s, nil
}

// Save writes the state atomically: encode to a temp file in the same
// directory, then rename over the state file.
func (s *State) Save(dir string) (err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".league-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp state: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp state: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp state: %w", err)
	}
	if err = os.Rename(tmpName, filepath.Join(dir, stateFileName)); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

// Exists reports whether a league state file exists in dir.
func Exists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, stateFileName))
	return err == nil
}
