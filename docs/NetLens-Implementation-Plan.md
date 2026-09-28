# NetLens — High-Level Implementation Plan

## 1. Purpose

This document is the implementation blueprint for NetLens.

It is intentionally **high level**. It defines the major implementation areas (epics), system boundaries, data flow, contracts, dependencies, and implementation order.

An agentic coding assistant should use this document to decompose the project into concrete implementation tickets/tasks. Individual tickets should contain the detailed code-level work, tests, and acceptance criteria.

The implementation should optimize for:

- A working Phase 1 MVP as early as possible
- Clear separation between networking logic and presentation
- A stable API contract between Go and the Svelte frontend
- Cross-platform operation on Linux, macOS, and Windows
- A simple single-binary production distribution
- Incremental delivery, with each epic leaving the application in a usable state
- Minimal frontend complexity and dependency count

---

# 2. Product Architecture

NetLens is a local network discovery and lightweight traffic-inspection application.

The application consists of:

```text
┌──────────────────────────────────────────────────────────────┐
│                         NetLens Binary                       │
│                                                              │
│  ┌──────────────────┐        ┌────────────────────────────┐  │
│  │     Scanner      │        │          Capture           │  │
│  │                  │        │                            │  │
│  │ ARP              │        │ Packet capture             │  │
│  │ ICMP             │        │ BPF filtering              │  │
│  │ Reverse DNS      │        │ Protocol decoding          │  │
│  │ mDNS             │        │ Traffic aggregation        │  │
│  │ SSDP             │        │                            │  │
│  │ DHCP observation │        │                            │  │
│  │ Port probes      │        │                            │  │
│  │ Fingerprinting   │        │                            │  │
│  └────────┬─────────┘        └─────────────┬──────────────┘  │
│           │                                │                 │
│           └──────────────┬─────────────────┘                 │
│                          ▼                                   │
│                  ┌───────────────┐                           │
│                  │  Domain State │                           │
│                  │ / Device Model│                           │
│                  └───────┬───────┘                           │
│                          │                                   │
│              ┌───────────┴───────────┐                       │
│              ▼                       ▼                       │
│        ┌───────────┐          ┌────────────┐                │
│        │ REST API  │          │ WebSocket  │                │
│        └─────┬─────┘          └─────┬──────┘                │
│              │                      │                        │
│              └──────────┬───────────┘                        │
│                         ▼                                    │
│                Embedded SvelteKit UI                        │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

The Go backend owns:

- Network discovery
- Packet capture
- Device identification
- Network state
- Scan lifecycle
- Traffic aggregation
- Persistence
- API contracts
- Application configuration

The frontend owns:

- Presentation
- User interaction
- Local UI state
- Filtering/sorting/display concerns
- Real-time rendering
- Starting and monitoring backend operations

The frontend must **not** duplicate networking/business logic.

---

# 3. Technology Baseline

## Backend

- Go
- `net/http` or `chi` for HTTP routing
- `github.com/google/gopacket`
- `gopacket/pcap`
- `golang.org/x/net/icmp`
- `golang.org/x/net/ipv4`
- `golang.org/x/net/ipv6`
- mDNS library such as `github.com/hashicorp/mdns` or `github.com/grandcat/zeroconf`
- SSDP library such as `github.com/koron/go-ssdp`
- OUI registry stored locally
- WebSocket implementation such as `nhooyr.io/websocket` or `gorilla/websocket`
- `golang.org/x/sync/errgroup`
- SQLite only when persistent scan history is introduced

## Frontend

- SvelteKit
- Svelte 5
- TypeScript
- Tailwind CSS
- Native `fetch`
- Native browser WebSocket API

Avoid additional frontend dependencies unless a concrete requirement justifies them.

Potential later dependencies:

- Virtualized table implementation if device counts require it
- D3 or a Svelte-compatible graph library for topology visualization
- Small UI component library if the application grows enough to justify one

## Distribution

- SvelteKit frontend built during release
- Frontend assets embedded into Go using `embed`
- Single Go binary distributed to users
- Npcap on Windows
- libpcap on Linux/macOS

---

# 4. Core Domain Model

Before implementing individual discovery mechanisms, establish a canonical internal representation of a network device.

A device should be identifiable primarily through stable network identifiers such as MAC address, while allowing IP addresses, hostnames, vendors, services, and discovery observations to change over time.

The domain model should support at least:

- Device identity
- MAC address
- One or more IP addresses
- IPv4/IPv6 information where available
- Hostname
- Vendor
- Device type/category
- Discovery sources
- Last-seen timestamp
- First-seen timestamp
- Reachability state
- Open services
- Service/banner information
- Interface/network context
- Traffic counters when capture is active

The model should also distinguish between:

- Facts directly observed by a discovery mechanism
- Derived/inferred information
- Current state
- Historical observations

The exact Go structs and JSON representation should be established as part of the implementation and kept stable once the frontend starts depending on them.

---

# 5. Event Model

Real-time changes should be represented by a small, explicit event protocol.

The backend should emit events for important state changes rather than forcing the frontend to poll constantly.

Conceptually:

```text
scan_started
scan_progress
device_discovered
device_updated
device_removed
scan_completed
scan_failed
capture_started
capture_stopped
traffic_update
error
```

Events should contain:

- Event type
- Timestamp where useful
- Related operation/scan ID where applicable
- Relevant entity payload

Example conceptual event:

```json
{
  "type": "device_discovered",
  "scanId": "scan-123",
  "device": {
    "id": "...",
    "ipAddresses": ["192.168.1.20"],
    "mac": "AA:BB:CC:DD:EE:FF",
    "vendor": "Example",
    "hostname": "printer.local"
  }
}
```

The exact schema should be defined once and reused by the backend and frontend.

---

# 6. REST API and WebSocket Boundary

The API is the primary contract between backend and frontend.

REST should be used for:

- Reading current state
- Starting operations
- Stopping operations
- Requesting device details
- Requesting exports
- Reading configuration/capabilities

WebSocket should be used for:

- Scan progress
- Newly discovered devices
- Device updates
- Capture state
- Live traffic updates
- Other events where polling would be inefficient

The frontend should never reach directly into scanner/capture internals.

Conceptual API:

```text
GET    /api/status
GET    /api/interfaces
GET    /api/network
GET    /api/devices
GET    /api/devices/:id

POST   /api/scans
GET    /api/scans/:id
POST   /api/scans/:id/cancel

POST   /api/capture/start
POST   /api/capture/stop
GET    /api/capture/status

GET    /api/devices/:id/traffic

GET    /api/export/devices.csv
GET    /api/export/devices.json

WS     /api/events
```

This is conceptual rather than a final endpoint list. The implementation agent should refine the API while preserving the separation of responsibilities.

The API should expose capabilities and errors clearly enough for the frontend to behave correctly on different operating systems and when packet capture is unavailable.

---

# 7. Epic 1 — Repository and Development Foundation

Establish the project structure and development workflow before implementing networking functionality.

Expected structure should roughly separate:

```text
/backend or Go packages
  scanner
  capture
  api
  domain
  discovery
  fingerprint
  storage
  config

/frontend
  SvelteKit application
```

The exact repository layout can be decided by the implementation agent.

This epic should establish:

- Go module
- SvelteKit application
- TypeScript
- Tailwind
- Local development commands
- Formatting
- Linting
- Unit test setup
- Basic integration test approach
- Frontend/backend development configuration
- Environment/configuration handling
- Build conventions
- CI foundations if appropriate

At the end of this epic, both backend and frontend should build independently and run locally.

---

# 8. Epic 2 — Application Shell and Embedded Frontend

Create the minimal application shell before building feature screens.

Backend responsibilities:

- Start HTTP server
- Serve API routes
- Serve a health/status endpoint
- Serve the frontend
- Support development mode where frontend can run independently
- Establish the production embedding mechanism

Frontend responsibilities:

- SvelteKit application shell
- Global layout
- Navigation
- Basic page routing
- Loading/error handling foundation
- API client abstraction
- WebSocket client abstraction
- Basic visual design system

The frontend should be able to connect to a backend and display basic server status before scanner functionality exists.

---

# 9. Epic 3 — Network and Interface Detection

Implement the infrastructure required to understand the local machine's networking environment.

The backend should identify:

- Available network interfaces
- Interface addresses
- IPv4 subnets
- IPv6 information where useful
- Likely local/default network
- Interface capabilities relevant to capture

This information should become available through the API.

The UI should provide enough information for the user to understand which interface/subnet NetLens is operating against.

Important considerations:

- Multiple interfaces
- VPN interfaces
- Virtual interfaces
- Docker/VM interfaces
- Loopback
- Wi-Fi/Ethernet
- Platform-specific behavior

Do not assume that the first interface returned by the OS is the correct LAN interface.

---

# 10. Epic 4 — Device Domain and State Management

Implement the central device inventory/state manager.

It should provide a single authoritative state model that all discovery mechanisms update.

Responsibilities:

- Create/update devices
- Merge observations from multiple discovery mechanisms
- Associate IP addresses with devices
- Track timestamps
- Track discovery sources
- Maintain service information
- Resolve conflicting observations sensibly
- Expose snapshots to the API
- Emit events when state changes

The scanner mechanisms should not each maintain independent copies of device state.

The domain/state layer should be concurrency-safe.

---

# 11. Epic 5 — Active ARP Discovery

Implement the primary LAN discovery mechanism.

The ARP scanner should:

1. Determine the target IPv4 subnet.
2. Generate ARP requests for appropriate addresses.
3. Send requests through the selected interface.
4. Listen for replies.
5. Convert responses into device observations.
6. Update the central device state.
7. Emit real-time discovery events.

The implementation should support:

- Concurrent requests where appropriate
- Timeouts
- Cancellation
- Rate limiting
- Interface selection
- Error handling
- Progress reporting

Avoid shelling out to external tools.

ARP discovery should be the first complete end-to-end discovery path.

---

# 12. Epic 6 — ICMP Discovery and Reverse DNS

Add complementary active discovery.

ICMP:

- Probe relevant addresses
- Record successful responses
- Associate results with existing device records
- Avoid treating ICMP failure as proof that a device does not exist

Reverse DNS:

- Attempt hostname resolution for discovered IPs
- Use bounded timeouts
- Avoid blocking the entire scan on slow DNS
- Store hostname observations separately from stronger identity signals

These mechanisms should enrich the existing inventory rather than create duplicate devices.

---

# 13. Epic 7 — Passive Discovery

Implement short passive observation windows for discovery signals.

Target protocols:

- ARP
- mDNS
- SSDP
- DHCP where observable

The passive discovery subsystem should:

- Select/open the appropriate capture interface
- Listen for relevant traffic
- Decode only required fields
- Convert observations into device-state updates
- Attribute observations to devices
- Emit device events
- Stop cleanly after the configured observation window

Passive discovery should be independently testable from the active scanner.

It should also be possible to run passive discovery without starting an active scan.

---

# 14. Epic 8 — Vendor Identification

Implement local MAC OUI lookup.

Requirements:

- Store the OUI data locally
- Load it at application startup or through an efficient lookup structure
- Normalize MAC addresses
- Resolve vendor from OUI
- Handle unknown/private/randomized MACs gracefully
- Avoid network access for normal vendor lookup

The vendor result is an enrichment, not a guaranteed device identity.

Randomized MAC addresses and locally administered addresses should be represented correctly rather than forcing an incorrect vendor.

---

# 15. Epic 9 — Service Discovery and Lightweight Fingerprinting

Implement lightweight service identification.

The objective is to answer:

> "What services does this device appear to expose?"

This is not intended to become a vulnerability scanner.

Start with a small set of common TCP ports.

For selected open ports:

- Establish safe connections
- Apply strict timeouts
- Avoid destructive interaction
- Capture simple banners where appropriate
- HTTP: optionally retrieve enough information to identify the server/title safely
- SSH: extract version string where available

Store:

- Port
- Protocol
- Service guess
- Banner/title where available
- Timestamp
- Confidence if useful

The scanner should remain conservative and configurable.

---

# 16. Epic 10 — Phase 1 Scan Orchestration

Combine all Phase 1 discovery mechanisms into a coherent scan lifecycle.

A scan should have:

```text
created
  ↓
initializing
  ↓
discovering
  ↓
enriching
  ↓
completed
```

With failure/cancellation paths.

The orchestrator should coordinate:

- Network detection
- ARP
- ICMP
- Reverse DNS
- Passive discovery
- OUI lookup
- Service probing
- Fingerprinting

Use concurrency where beneficial, but preserve predictable resource usage.

The orchestrator should produce progress events.

A user should be able to:

- Start a scan
- See progress
- See devices appear while the scan runs
- Cancel a scan
- See completion/failure
- Inspect the resulting inventory

This epic produces the first complete NetLens MVP.

---

# 17. Epic 11 — Phase 1 Frontend

Build the primary user experience around the scan lifecycle.

Primary screens:

### Dashboard

Show:

- Current network/interface
- Scan state
- Device count
- Last scan time
- Scan action

### Device Inventory

Provide:

- Search
- Sorting
- Basic filtering
- Device status
- IP
- MAC
- Vendor
- Hostname
- Device type
- Services summary

The initial implementation should use standard Svelte components and a normal table/list.

Do not introduce virtualization unless actual data volumes justify it.

### Device Detail

Show:

- Identity
- Addresses
- Vendor
- Hostname
- Device type
- Discovery sources
- Services
- Last seen
- Relevant observations

### Real-Time Updates

Use the WebSocket event stream so that:

- Devices appear during scans
- Device changes are reflected immediately
- Scan progress updates without polling

The UI should handle:

- Loading
- Empty state
- Errors
- Permission/capture-driver errors
- Cancelled scans
- No devices found

---

# 18. Epic 12 — Phase 1 UX and Hardening

After the MVP works end-to-end, improve reliability and usability.

Areas:

- Clear first-run experience
- Interface selection
- Local subnet presentation
- Scan confirmation
- Permission/capability messaging
- Responsive layout
- Accessibility
- Error messages
- Consistent loading states
- Empty states
- Long-running operation handling
- Browser refresh/reconnect behavior
- WebSocket reconnect behavior

The goal is to make the application feel like a finished tool rather than a developer prototype.

---

# 19. Epic 13 — Capture Infrastructure

Begin Phase 2.

Implement a dedicated capture subsystem independent from the scanner.

Responsibilities:

- Open selected interface
- Configure capture
- Apply BPF filter
- Start/stop capture
- Handle capture errors
- Produce decoded packet events
- Provide capture lifecycle state

Capture should be optional.

NetLens must remain useful when packet capture is unavailable.

This subsystem should not expose raw packet streams directly to the frontend.

Instead, it should produce normalized events and aggregated traffic information.

---

# 20. Epic 14 — Lightweight Packet Decoding

Decode only the deliberately supported protocol set:

- Ethernet where required
- ARP
- ICMP
- IPv4/IPv6
- TCP
- UDP
- DNS
- DHCP
- TLS ClientHello / SNI

The decoder should extract only information needed by the product.

Examples:

```text
source device
destination IP
destination hostname/SNI when available
protocol
source port
destination port
packet size
timestamp
```

Do not implement full packet-forensics functionality.

Do not build a general-purpose protocol dissector.

---

# 21. Epic 15 — Traffic Aggregation

Transform packet events into useful device-level information.

Aggregate:

- Packets sent
- Packets received
- Bytes sent
- Bytes received
- Recent destinations
- Recent protocols
- Recent ports
- Recent hostnames/SNI
- Time windows

The aggregation layer should prevent the frontend from receiving every packet when that is unnecessary.

For example:

```text
raw packets
     ↓
decoder
     ↓
device attribution
     ↓
traffic aggregation
     ↓
periodic traffic updates
     ↓
WebSocket
     ↓
frontend
```

The system should have bounded memory usage.

---

# 22. Epic 16 — Phase 2 Frontend

Add a traffic-oriented UI without turning NetLens into Wireshark.

Potential views:

### Device Traffic

For a selected device:

- Recent destinations
- Protocols
- Ports
- Packets
- Bytes
- Recent hostnames/SNI

### Live Traffic

Provide a simplified real-time view:

```text
Device       Destination       Protocol   Activity
MacBook      github.com        HTTPS      ↑ 42 KB
NAS          192.168.1.1       DNS        ↑ 1 KB
RaspberryPi  pool.ntp.org      NTP        ↑ 128 B
```

Provide filtering by:

- Device
- Host
- Port
- Protocol

Use WebSockets for live updates.

Throttle UI updates if packet activity is high.

---

# 23. Epic 17 — Persistence

Add optional SQLite-backed persistence.

Persistence should be optional and should not complicate the default experience.

Persist information needed for:

- Previous scan results
- Scan timestamps
- Device history
- Device first/last seen
- Change detection
- Historical reports

Do not persist raw packet streams.

The storage layer should be behind an interface so that in-memory operation remains available.

---

# 24. Epic 18 — Scan History and Change Detection

Use persisted device state to answer:

> "What changed since the previous scan?"

Detect:

- New devices
- Missing devices
- IP changes
- Hostname changes
- Vendor/type changes
- Service changes where meaningful

Expose a concise change view.

Example:

```text
Network changes

+ 192.168.1.37  Espressif
+ 192.168.1.41  Apple

- 192.168.1.22  Samsung

Changed:
192.168.1.10
  IP: 192.168.1.10 → 192.168.1.11
```

Changes should be based on actual observations and should avoid noisy false positives.

---

# 25. Epic 19 — Reporting and Export

Implement:

- CSV export
- JSON export

Exports should be generated from the canonical device inventory.

Do not make the frontend responsible for reconstructing export data.

The backend should provide downloadable export endpoints.

The export format should be documented and reasonably stable.

---

# 26. Epic 20 — Topology Visualization

Implement the network topology view after the underlying device and relationship data is mature.

The topology model should represent useful relationships such as:

```text
Gateway
   │
   ├── Device A
   ├── Device B
   ├── Device C
   └── Device D
```

Do not imply network relationships that the system cannot actually observe.

The visualization should distinguish:

- Known facts
- Inferred relationships
- Unknown relationships

Use a graph library only when the underlying topology data is ready.

The graph is a visualization layer, not a source of network truth.

---

# 27. Epic 21 — Cross-Platform Support

Validate and harden the application on:

- Linux
- macOS
- Windows

Areas requiring platform-specific testing:

- Interface enumeration
- Subnet detection
- Raw ICMP behavior
- ARP behavior
- Packet capture
- Npcap/libpcap availability
- Permissions
- Firewall interaction
- Interface naming
- IPv6 behavior

The application should report actionable errors such as:

- Capture driver missing
- Insufficient privileges
- Interface unavailable
- Capture unsupported
- Permission denied

Avoid silently falling back to behavior that produces misleading results.

---

# 28. Epic 22 — Security, Safety, and Authorization UX

NetLens performs active network operations.

Implement safeguards around scanning scope.

Default behavior:

- Detect the local subnet
- Default scans to that subnet
- Clearly show the target network before scanning
- Provide a first-run authorization/consent notice
- Require explicit opt-in before scanning outside the detected local subnet

The backend should validate scan targets rather than relying solely on frontend restrictions.

Do not treat frontend validation as a security boundary.

Network probing should have reasonable limits for:

- Target ranges
- Concurrency
- Timeouts
- Request rates

---

# 29. Epic 23 — Testing Strategy

Testing should exist at multiple levels.

## Unit tests

Cover:

- Device identity/merge logic
- MAC normalization
- OUI lookup
- Network calculations
- Event serialization
- Packet decoding
- Traffic aggregation
- Change detection
- Configuration

## Integration tests

Cover:

- REST API
- WebSocket event delivery
- Scan lifecycle
- Storage
- Export generation
- Backend/frontend contract

## Network tests

Where practical, use controlled fixtures or test environments rather than relying exclusively on the developer's LAN.

Test:

- ARP parsing
- Packet decoding
- mDNS/SSDP/DHCP parsing
- TCP banner parsing
- TLS SNI extraction

## Frontend tests

Cover important UI behavior:

- Device rendering
- Filtering
- Sorting
- Scan lifecycle
- WebSocket updates
- Error states
- Device details

The exact testing framework is implementation-specific.

---

# 30. Epic 24 — Build, Packaging, and Release

Create a reproducible production build.

The release pipeline should conceptually be:

```text
checkout
   ↓
build frontend
   ↓
validate frontend
   ↓
copy/embed frontend assets
   ↓
run Go tests
   ↓
build Go binaries
   ↓
package releases
```

Produce platform-specific binaries where appropriate.

Document:

- Required capture driver
- Permissions
- Installation
- Starting NetLens
- Opening the web UI
- Supported platforms

The normal user experience should remain:

```text
install/run NetLens
        ↓
open browser
        ↓
scan LAN
```

No Node.js installation should be required for end users.

---

# 31. Epic 25 — Documentation and Demo

Create documentation that explains:

- What NetLens is
- What it is not
- Installation
- Platform requirements
- Packet capture requirements
- Basic usage
- Scanning behavior
- Privacy considerations
- Legal/ethical considerations
- Architecture
- API
- Development setup
- Release/build process

Create a polished demo showing:

1. NetLens starts
2. Local interface/subnet is detected
3. Scan begins
4. Devices appear live
5. Device details are inspected
6. Traffic view is opened
7. A useful network observation is made

The project should be understandable from the GitHub repository without reading the source code.

---

# 32. Recommended Implementation Order

The project should be implemented vertically rather than completing all backend work first and frontend work later.

Recommended order:

```text
1. Repository/development foundation
        ↓
2. Backend HTTP shell + SvelteKit shell
        ↓
3. API/WebSocket contract
        ↓
4. Network/interface detection
        ↓
5. Device domain/state manager
        ↓
6. ARP discovery
        ↓
7. End-to-end scan + live frontend inventory
        ↓
8. ICMP + reverse DNS
        ↓
9. OUI lookup
        ↓
10. Passive discovery
        ↓
11. Service fingerprinting
        ↓
12. Phase 1 hardening/release
        ↓
13. Capture infrastructure
        ↓
14. Packet decoding
        ↓
15. Traffic aggregation
        ↓
16. Phase 2 frontend
        ↓
17. Persistence
        ↓
18. Change detection
        ↓
19. Export
        ↓
20. Topology
        ↓
21. Cross-platform/release hardening
```

The key milestone is **end-to-end Phase 1**.

At that point the application should already be useful:

```text
Start NetLens
     ↓
Detect local network
     ↓
Scan
     ↓
Discover devices
     ↓
Identify devices
     ↓
See results live in browser
```

Everything after that should build on this foundation.

---

# 33. Agentic Development Guidelines

The implementation agent should follow these principles when converting this document into tickets.

## Work vertically

Prefer tickets that produce observable behavior.

Good:

> Implement ARP discovery and expose discovered devices through the API and WebSocket, including tests.

Less useful:

> Create scanner package.

## Keep contracts explicit

Before implementing a frontend feature, define the corresponding backend API/event contract.

Avoid frontend code depending on undocumented backend behavior.

## Avoid speculative infrastructure

Do not introduce:

- Redux-like state management
- Complex dependency injection
- Message brokers
- Microservices
- External databases
- Cloud infrastructure

unless a concrete requirement emerges.

NetLens is intentionally a local, single-binary application.

## Keep networking modules isolated

Scanner mechanisms should be independently testable.

Capture should remain separate from scanning.

The API should depend on domain/application services rather than directly manipulating packet-capture code.

## Prefer boring solutions

Use standard Go libraries and straightforward concurrency patterns wherever practical.

For the frontend, prefer Svelte's native capabilities before adding dependencies.

## Preserve cancellation and bounded resources

Network operations must support:

- Context cancellation
- Timeouts
- Bounded concurrency
- Bounded memory
- Clean shutdown

Long-running operations must not leak goroutines, sockets, capture handles, or WebSocket connections.

## Make failures visible

If a platform cannot perform a capability, expose a clear capability/error state.

Do not silently report incomplete discovery as a complete scan.

## Keep the MVP narrow

Phase 1 should be considered complete when the product can reliably answer:

> "What devices are currently on this LAN, and what do we know about them?"

Do not delay Phase 1 for topology, historical storage, advanced packet decoding, or elaborate UI.

---

# 34. Definition of a Successful MVP

The first production-quality milestone is achieved when a user can:

1. Start the NetLens binary.
2. Open the local web UI.
3. See the detected network/interface.
4. Start a scan.
5. Watch devices appear while the scan runs.
6. Search and sort the inventory.
7. Open a device detail view.
8. See IP, MAC, vendor, hostname, and discovered services where available.
9. Understand when the scan is running, completed, cancelled, or failed.
10. Run the application without Node.js or other frontend tooling installed.
11. Run it on the supported target OS with the required packet-capture dependency.
12. Build and test the project reproducibly from the repository.

The MVP should feel like a complete small product even before Phase 2 exists.

---

# 35. Final Architectural Principle

NetLens should remain a **local-first, backend-heavy, single-binary application**.

The central flow is:

```text
             ┌──────────────────┐
             │ Physical Network │
             └────────┬─────────┘
                      │
          ┌───────────┴───────────┐
          ▼                       ▼
   Active Discovery        Passive/Capture
          │                       │
          └───────────┬───────────┘
                      ▼
              Domain / Device State
                      │
             ┌────────┴────────┐
             ▼                 ▼
          REST API         WebSocket
             │                 │
             └────────┬────────┘
                      ▼
                 SvelteKit UI
                      │
                      ▼
                   Browser
```

The Go backend is the authoritative source of network truth.

The SvelteKit frontend is a thin, responsive, real-time interface over that backend.

The system should be built incrementally, with **Phase 1 delivering a useful product before Phase 2 and Phase 3 add complexity**.
