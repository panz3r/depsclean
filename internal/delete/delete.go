package delete

import "github.com/panz3r/npclean/internal/model"

// Deleter handles safe removal of discovered target directories.
// Implementation is deferred to a later phase.
type Deleter struct {
	DryRun bool
}

// Delete removes the directory described by result, or simulates removal if DryRun is set.
func (d *Deleter) Delete(result model.Result) error {
	return nil
}
