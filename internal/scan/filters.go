package scan

// Filter defines a predicate that can exclude paths from scanning.
// Implementation is deferred to a later phase.
type Filter func(path string) bool
