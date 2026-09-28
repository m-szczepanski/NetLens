// Package scanner implements active device discovery mechanisms (ARP sweep,
// ICMP echo) that observe the local network and feed results into the domain
// state. Discovery mechanisms must not keep independent copies of device
// state and must support cancellation, timeouts, and bounded concurrency.
package scanner
