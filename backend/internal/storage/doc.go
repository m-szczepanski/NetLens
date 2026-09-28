// Package storage persists application state when needed: in-memory by
// default, optional embedded SQLite for scan history across runs. Added only
// once persistence is actually required.
package storage
