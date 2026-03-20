# ADR: P2P Peer Set Duplicate Protection

## Context

`set.Add()` in `tm2/pkg/p2p/set.go` unconditionally increments the `inbound`
or `outbound` counter without checking if the peer ID already exists. The map
overwrites duplicates silently, but the counter always increments.

This means N calls to `Add()` with the same peer ID produce 1 map entry but a
counter of N. When all connections close, only one `Remove()` succeeds
(decrementing once), leaving a permanent ghost counter of N-1.

The inbound accept loop (`runAcceptLoop()`) gates solely on `NumInbound()`,
and the transport deduplicates by IP:port — not peer ID. So the same keypair
from different IPs inflates the counter with no guard. With
`maxInboundPeers = 40`, this permanently exhausts all inbound slots.

## Decision

Two complementary fixes:

### 1. Make `set.Add()` idempotent (root cause)

Check if the peer ID exists before touching counters. If it does, only adjust
counters when the direction (inbound/outbound) changed:

```go
if existing, exists := s.peers[peer.ID()]; exists {
    if existing.IsOutbound() && !peer.IsOutbound() {
        s.outbound -= 1
        s.inbound += 1
    } else if !existing.IsOutbound() && peer.IsOutbound() {
        s.inbound -= 1
        s.outbound += 1
    }
    s.peers[peer.ID()] = peer
    return
}
```

The direction-change branch should never trigger in practice (Fix 2 blocks
duplicates before they reach `Add()`), but it keeps the data structure
self-consistent in isolation.

### 2. Reject duplicate peer IDs in `runAcceptLoop()` (defense in depth)

Add `peers.Has(p.ID())` before `addPeer()`. This blocks the primary vector
and avoids wasted work (peer start, reactor init).

## Alternatives considered

**A. Ignore direction change on duplicate** — rejected because if a duplicate
with a different direction reaches `Add()`, `Remove()` would later decrement
the wrong counter (uint64 underflow).

**B. Reject duplicates in `Add()` (return bool)** — rejected because the peer
is already started before `Add()` is called; the caller cleanup would require
interface changes.

**C. Only apply the switch-level guard** — rejected because `set.Add()` would
still violate its own invariants. Any future caller without a prior `Has()`
check would reintroduce the bug.

## Consequences

- Ghost counter inflation blocked at two layers
- `set` counters always match map contents
- 5 new regression tests, no regressions on existing tests
- `set.Add()` slightly more complex due to direction-change handling
