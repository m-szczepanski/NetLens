# A browser-based tool for people who build and configure networks, focused on knowing everything connected to a LAN - not a Wireshark clone.

## Vision

Network installers, IT admins, and self-hosters constantly need to answer one question: "what's actually on this network?" Existing tools split the job awkwardly — Wireshark shows you every byte but nothing about devices; nmap tells you about hosts but lives in a terminal; router admin pages are vendor-locked and ugly. NetLens is a single Go binary that scans a LAN, builds a live device inventory with vendor/type/hostname info, and serves a clean web UI to browse it — with a little packet-level insight layered on top for troubleshooting, not full protocol forensics.

## Goals

- Discover every device on a local subnet, actively and passively, in seconds
- Identify what each device is (vendor, likely device type, open services) without manual digging
- Present it in a fast, minimal, non-technical-friendly web UI
- Offer enough live traffic visibility to answer "is this device talking, and to what" - without trying to be a packet forensics suite
- Ship as one static binary, no install dependencies beyond a capture driver

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

Backend (Go): two loosely-coupled modules - scanner (device discovery, the primary feature) and capture (optional, Phase 2 packet insight) - behind a REST + WebSocket API.
Frontend: single-page web app, WebSocket for live device/packet updates, REST for on-demand scans and exports.
Storage: in-memory by default; optional embedded SQLite for scan history across runs.
Distribution: single static Go binary with the frontend embedded via embed, so there's nothing to install except a packet-capture driver (Npcap on Windows; libpcap is usually already present on Linux/macOS).

### Libraries

#### Go backend

- gopacket + gopacket/pcap - capture and the small set of protocol decodes needed for Phase 2
- golang.org/x/net/icmp, golang.org/x/net/ipv4/ipv6 - raw ICMP echo without shelling out to ping
- ARP request/reply construction and parsing via gopacket/layers
- github.com/hashicorp/mdns or github.com/grandcat/zeroconf - mDNS/DNS-SD discovery
- github.com/koron/go-ssdp - SSDP/UPnP discovery
- MAC vendor lookup - a local copy of the IEEE OUI registry (CSV) plus a small lookup wrapper, e.g. github.com/klauspost/oui
- nhooyr.io/websocket or gorilla/websocket - pushing live updates to the browser
- chi or echo (or plain net/http) for routing
- golang.org/x/sync/errgroup - coordinating concurrent scan workers cleanly

#### Frontend

- React + TypeScript (or Svelte if you want something lighter)
- Tailwind CSS for the minimal look
- TanStack Table + TanStack Virtual (or react-window) - device table stays smooth even with hundreds of rows
- react-force-graph or D3 force layout - Phase 3 topology map
- Native browser WebSocket API - no client library needed
  
#### Dev/ops

Go's built-in cross-compilation; note that libpcap/Npcap must exist at runtime on the target OS - document the install step per platform in the README, and consider bundling the Npcap installer for Windows releases

### Background Reading

RFC 826 - Address Resolution Protocol (ARP)
RFC 792 - Internet Control Message Protocol (ICMP)
RFC 6762 - Multicast DNS (mDNS)
RFC 6763 - DNS-Based Service Discovery (DNS-SD)
RFC 2131 - DHCP (useful for passive discovery signals)
UPnP Forum device architecture docs - SSDP has no single RFC, this is the reference
IEEE OUI registry - vendor lookup data source
gopacket documentation and its examples/ directory
BPF (Berkeley Packet Filter) syntax reference - used for capture filters
Npcap docs (Windows) and libpcap docs (Linux/macOS)
nmap's host discovery chapter - good conceptual reference even without using nmap itself

## Legal/ethical Note
Active scanning (ARP sweeps, ICMP, port probing) of a network you don't own or administer can violate laws or ISP terms of service in some jurisdictions. Worth a first-run consent screen, defaulting the tool to the local subnet only, with an explicit opt-in before scanning anything wider.

---

netlens - a lightweight network discovery tool. Scans your LAN, identifies every connected device (vendor, type, open services), and shows it all in a clean web UI - with just enough live packet insight to troubleshoot, without trying to be Wireshark.

# Rough Roadmap
Weeks 1–2: Phase 1 - ARP/ICMP sweep, OUI lookup, device table UI
Week 3: passive discovery (mDNS/SSDP/DHCP sniffing), banner grabs
Weeks 4–5: Phase 2 - capture module, small protocol decode set, per-device traffic counters
Weeks 6+: Phase 3 - topology graph, CSV/JSON export, scheduled re-scans with diffing
