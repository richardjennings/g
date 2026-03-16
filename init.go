package g

import (
	"fmt"
	"os"
)

// Init creates a new git repository and returns the Repository.
func Init(opts ...Opt) (*Repository, error) {
	r, err := Open(opts...)
	if err != nil {
		return nil, err
	}
	path := r.gitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}
	for _, v := range []string{
		r.objectPath(),
		r.refsDirectory(),
		r.refsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", v, err)
		}
	}
	if err := r.updateHead(r.defaultBranch()); err != nil {
		return nil, err
	}
	return r, nil
}

// initRepo is a free-function shim for tests that still use Configure() + Init().
func initRepo() error {
	path := defaultRepo.gitPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, v := range []string{
		defaultRepo.objectPath(),
		defaultRepo.refsDirectory(),
		defaultRepo.refsHeadsDirectory(),
	} {
		if err := os.MkdirAll(v, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", v, err)
		}
	}
	return defaultRepo.updateHead(defaultRepo.defaultBranch())
}
