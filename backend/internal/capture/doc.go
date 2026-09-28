// Package capture implements optional packet capture (Phase 2): opening a
// capture handle on a chosen interface, decoding the small protocol set, and
// reporting per-device traffic counters. Capture stays separate from
// scanning. Capture handles must be closed cleanly and never leak.
package capture
