# NetLens

A browser-based tool for people who build and configure networks, focused on knowing everything connected to a LAN - not a Wireshark clone.

## Vision

Network installers, IT admins, and self-hosters constantly need to answer one question: "what's actually on this network?" Existing tools split the job awkwardly — Wireshark shows you every byte but nothing about devices; nmap tells you about hosts but lives in a terminal; router admin pages are vendor-locked and ugly.

NetLens is a single Go binary that scans a LAN, builds a live device inventory with vendor/type/hostname info, and serves a clean web UI to browse it — with a little packet-level insight layered on top for troubleshooting, not full protocol forensics.

## Goals

- Discover every device on a local subnet, actively and passively, in seconds
- Identify what each device is (vendor, likely device type, open services) without manual digging
- Present it in a fast, minimal, non-technical-friendly web UI
- Offer enough live traffic visibility to answer "is this device talking, and to what" - without trying to be a packet forensics suite
- Ship as one static Go binary, with no install dependencies beyond a packet-capture driver

## Non-goals

- Reproducing Wireshark's ~3,000 protocol dissectors or its display-filter grammar
- Full stream reconstruction, file carving, or long-term full-traffic (pcap firehose) recording
- Deep packet inspection for security/forensic investigation
- Being a general-purpose port/vulnerability scanner (that's nmap/Nessus territory)

## Target Users

1. Network installers/integrators setting up or auditing SMB and home networks
2. IT admins who need a quick inventory of what's plugged into a LAN
3. Homelab/self-hosting hobbyists

## Feature Scope, Phased

### Phase 1 - Core Discovery (MVP)

1. Active discovery: ARP sweep of the local subnet, ICMP echo sweep, reverse DNS
2. Passive discovery: short listen window on broadcast/multicast traffic (ARP requests, mDNS, SSDP, DHCP) to catch devices without probing them
3. Vendor ID via MAC OUI lookup
4. Lightweight fingerprinting: common TCP ports open, safe banner grabs (HTTP title, SSH version string)
5. Web UI: sortable/searchable device table + detail panel per device

### Phase 2 - Light Packet Insight

1. Live capture on a chosen interface, filterable by host/port/protocol
2. Decode a deliberately small protocol set: ARP, ICMP, DNS, DHCP, TCP/UDP headers, TLS ClientHello SNI (enough to say "this device is reaching out to example.com")
3. Per-device traffic counters (recent packets/bytes), not full stream reconstruction

### Phase 3 - Topology & Reporting

1. Visual network map: gateway/subnet/devices as a graph
2. Export inventory as CSV/JSON
3. Periodic re-scans with diffing - "what changed since last scan," new/missing devices flagged

## Architecture

### Backend

**Go**: two loosely-coupled modules:

- `scanner` - device discovery, the primary feature
- `capture` - optional Phase 2 packet insight

The backend exposes:

- **REST API** for scans, device inventory, on-demand operations, and exports
- **WebSocket API** for live device, scan, and packet updates

The backend is the source of truth for network state and business logic. The frontend should remain a thin presentation and interaction layer rather than duplicating discovery or networking logic.

### Frontend

**SvelteKit + Svelte 5 + TypeScript + Tailwind CSS.**

The frontend is intentionally lightweight and dependency-minimal. It is a browser application that:

- Displays the device inventory
- Provides search, sorting, filtering, and device detail views
- Starts and monitors scans through the REST API
- Receives live updates through the native browser WebSocket API
- Displays light packet/traffic information in Phase 2
- Provides topology visualization and reporting UI in Phase 3

Avoid adding state-management, table, virtualization, or visualization libraries unless the application actually needs them. Prefer Svelte's built-in reactivity and the browser's native APIs first.

For example, the initial device table should be implemented with normal Svelte components. Introduce virtualization only if real-world device counts make it necessary.

### Frontend/Backend Boundary

The frontend communicates with the Go backend through a clearly defined API contract.

Conceptually:

```text
                 ┌─────────────────────────┐
                 │       Go Backend        │
                 │                         │
                 │ scanner │ capture       │
                 │         │               │
                 │ REST API│ WebSocket     │
                 └─────────┬───────────────┘
                           │
                    API / events
                           │
                 ┌─────────▼───────────────┐
                 │      SvelteKit UI       │
                 │                         │
                 │ pages │ components      │
                 │ state │ API client      │
                 └─────────────────────────┘
```

The API contract should be documented and treated as the source of truth for frontend development. The frontend should not invent endpoints or reproduce backend business logic.

### Storage

In-memory by default; optional embedded SQLite for scan history across runs.

### Distribution

Production NetLens is distributed as a **single Go binary** with the built SvelteKit frontend embedded using Go's `embed` package.

The production flow is:

```text
SvelteKit source
      │
      │ npm run build
      ▼
static frontend assets
      │
      │ go:embed
      ▼
single Go binary
      │
      ├── REST API
      ├── WebSocket API
      └── embedded web UI
```

There is no Node.js or Svelte runtime dependency on the target machine.

The only expected external runtime dependency is a packet-capture driver:

- Npcap on Windows
- libpcap, usually already present, on Linux/macOS

The frontend can also be deployed independently as a static web application for development, demos, documentation, or a hosted UI. Netlify and similar platforms support SvelteKit directly. However, a hosted frontend does **not** perform LAN scanning itself; the Go backend must run somewhere with network access to the LAN being inspected.

## Libraries

### Go backend

- `github.com/google/gopacket` + `gopacket/pcap` - capture and the small set of protocol decodes needed for Phase 2
- `golang.org/x/net/icmp`, `golang.org/x/net/ipv4/ipv6` - raw ICMP echo without shelling out to ping
- ARP request/reply construction and parsing via `gopacket/layers`
- `github.com/hashicorp/mdns` or `github.com/grandcat/zeroconf` - mDNS/DNS-SD discovery
- `github.com/koron/go-ssdp` - SSDP/UPnP discovery
- MAC vendor lookup - a local copy of the IEEE OUI registry (CSV) plus a small lookup wrapper, e.g. `github.com/klauspost/oui`
- `nhooyr.io/websocket` or `gorilla/websocket` - pushing live updates to the browser
- `chi` or `echo` (or plain `net/http`) for routing
- `golang.org/x/sync/errgroup` - coordinating concurrent scan workers cleanly

### Frontend

- **SvelteKit**
- **Svelte 5**
- **TypeScript**
- **Tailwind CSS**
- Native browser **WebSocket API**
- Native browser `fetch` API for REST calls

Potential future dependencies, only when justified by actual requirements:

- Table virtualization library if device counts require it
- D3 or a Svelte-compatible graph library for the Phase 3 topology map
- A small UI/component library if the application grows enough to benefit from one

### Dev/ops

- Node.js/npm (or equivalent package manager) is required only for frontend development and building the frontend
- Go's built-in cross-compilation is used for backend releases
- `libpcap`/Npcap must exist at runtime on the target OS
- Document the packet-capture driver installation step per platform in the README
- Consider bundling the Npcap installer for Windows releases

## Development

### Frontend

The frontend is developed independently as a SvelteKit application.

Typical development flow:

```bash
cd frontend
npm install
npm run dev
```

The SvelteKit development server provides fast local iteration. Configure the frontend API/WebSocket base URL so it can communicate with a locally running Go backend.

Before integrating the frontend into the Go binary:

```bash
npm run check
npm run build
```

### Production build

The release process should build the frontend first and then embed the generated static assets into the Go binary.

The exact build/embedding mechanism should be kept simple and reproducible, for example:

```text
1. Build SvelteKit frontend
2. Copy generated assets to the Go embed directory
3. Run Go tests
4. Build the Go binary
5. Package platform-specific release artifacts
```

## Hosting the Frontend Separately

SvelteKit is straightforward to deploy to platforms such as Netlify. Netlify supports SvelteKit directly and can automatically detect the framework and build configuration.

For NetLens, however, separate frontend hosting is primarily useful for:

- UI development
- public demos
- documentation
- a hosted frontend connecting to a separately reachable NetLens backend

It does not replace the Go backend, because LAN discovery and packet capture must happen in an environment with access to the target network.

For the normal self-hosted product, prefer the embedded-frontend architecture so users simply run the NetLens binary and open the provided local web address.

## Background Reading

- RFC 826 - Address Resolution Protocol (ARP)
- RFC 792 - Internet Control Message Protocol (ICMP)
- RFC 6762 - Multicast DNS (mDNS)
- RFC 6763 - DNS-Based Service Discovery (DNS-SD)
- RFC 2131 - DHCP (useful for passive discovery signals)
- UPnP Forum device architecture docs - SSDP has no single RFC, this is the reference
- IEEE OUI registry - vendor lookup data source
- gopacket documentation and its `examples/` directory
- BPF (Berkeley Packet Filter) syntax reference - used for capture filters
- Npcap docs (Windows) and libpcap docs (Linux/macOS)
- nmap's host discovery chapter - good conceptual reference even without using nmap itself

## Legal/ethical Note

Active scanning (ARP sweeps, ICMP, port probing) of a network you don't own or administer can violate laws or ISP terms of service in some jurisdictions.

The application should:

- Default to the local subnet
- Provide a clear first-run consent notice
- Require explicit opt-in before scanning networks outside the detected local subnet
- Clearly communicate that the user is responsible for having authorization to scan the target network

---

**netlens** - a lightweight network discovery tool. Scans your LAN, identifies every connected device (vendor, type, open services), and shows it all in a clean web UI - with just enough live packet insight to troubleshoot, without trying to be Wireshark.

## Rough Roadmap

- **Weeks 1–2:** Phase 1 - ARP/ICMP sweep, OUI lookup, device table UI
- **Week 3:** Passive discovery (mDNS/SSDP/DHCP sniffing), banner grabs
- **Weeks 4–5:** Phase 2 - capture module, small protocol decode set, per-device traffic counters
- **Weeks 6+:** Phase 3 - topology graph, CSV/JSON export, scheduled re-scans with diffing
