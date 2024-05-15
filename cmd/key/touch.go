package key

import (
	"errors"
	"fmt"
	"os"
)

// Test if I have write permission to a file at given path.
func touch(p string) (*os.File, error) {
	if info, err := os.Stat(p); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat: %w", err)
		}
	} else if info.IsDir() {
		return nil, fmt.Errorf("it must not a directory")
	}

	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0644)
	if err == nil {
		return f, nil
	}
	if errors.Is(err, os.ErrPermission) {
		return nil, err
	}

	return nil, fmt.Errorf("unknown err: %w", err)
}
