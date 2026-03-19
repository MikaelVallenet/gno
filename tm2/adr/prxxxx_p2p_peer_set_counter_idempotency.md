# ADR: P2P Peer Set Counter Idempotency

## Context

`set.Add()` in `tm2/pkg/p2p/set.go` unconditionally increments the `inbound`
or `outbound` counter every time a peer is added, without checking whether a
peer with the same ID already exists in the map. The Go map silently
overwrites duplicate keys, but the counter always increments, creating a
desynchronization between the actual peer count (`len(peers)`) and the
reported count (`NumInbound()` / `NumOutbound()`).

### The problem

When a peer ID is added N times (map overwrites to 1 entry, counter reaches
N), then removed, only the first `Remove()` succeeds (deletes the entry,
decrements by 1). The remaining N-1 calls find nothing and return without
decrementing. Final state: 0 actual peers, counter = N-1. This "ghost
counter" permanently inflates `NumInbound()`, blocking new inbound
connections until the node is restarted.

The inbound accept loop (`runAcceptLoop()` in `switch.go`) has no peer ID
deduplication — it only checks `NumInbound() >= maxInboundPeers`. The
transport layer deduplicates by `RemoteAddr().String()` (IP:port), not by
peer ID, so the same keypair connecting from different IPs bypasses all
duplicate detection.

With `maxInboundPeers = 40` (default), a single cycle of duplicate
connections can exhaust all inbound slots permanently.

### Code paths verified

- `set.Add()` (`set.go:27-40`): no existence check before counter increment
- `addPeer()` (`switch.go:704`): no `Has()` guard before `Add()`
- `runAcceptLoop()` (`switch.go:654`): gates only on `NumInbound()`, no peer
  ID deduplication
- `transport.processConn()` (`transport.go:241`): `activeConns` keyed by
  `RemoteAddr`, not peer ID
- `set.Remove()` (`set.go:72-92`): returns false if ID not found, does not
  decrement

## Decision

Two complementary fixes applied:

### 1. Make `set.Add()` idempotent (root cause fix)

Before incrementing a counter, check if the peer ID already exists in the map.
If it does, update the peer reference but only adjust counters if the direction
(inbound/outbound) changed. If the direction is the same, counters are
untouched.

```go
if existing, exists := s.peers[peer.ID()]; exists {
    // adjust counters only if direction changed
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

The direction-change branch ensures the counters always reflect what is
actually in the map. This should not trigger in practice (the switch-level
guard prevents duplicates from reaching `Add()`), but it makes the data
structure self-consistent in isolation.

### 2. Reject duplicate peer IDs in `runAcceptLoop()` (defense in depth)

Add a `peers.Has(p.ID())` check in the accept loop, before calling
`addPeer()`. This rejects duplicate inbound connections early, avoiding wasted
work (peer start, reactor initialization) and blocking the primary vector.

## Alternatives considered

### A. Skip counters on duplicate, ignore direction change

The simplest fix: if the peer ID exists, update the reference but never touch
counters regardless of direction.

**Rejected** because if a duplicate with a different direction somehow reaches
`Add()`, the counter would not match the map entry's direction. A subsequent
`Remove()` would decrement the wrong counter, causing a uint64 underflow.
Handling direction changes keeps the data structure self-consistent.

### B. Make `set.Add()` reject duplicates entirely (return bool, don't overwrite)

Change the signature to `Add(peer PeerConn) bool`, returning false if the
peer ID already exists, without updating the map or counters.

**Rejected** because in `addPeer()` the peer is already started
(`p.Start()` at line 696) before `Add()` is called (line 704). If `Add()`
rejects the peer, the caller must stop it and clean up reactors. This
requires changes to the `PeerSet` interface and all callers, and the
orphaned-peer cleanup is error-prone.

### C. Only apply the switch-level guard, leave `set.Add()` as-is

Rely solely on the `Has()` check in `runAcceptLoop()` to prevent duplicates.

**Rejected** because `set.Add()` would remain a broken data structure that
violates its own invariants. Any future caller that reaches `Add()` without
a prior `Has()` check would reintroduce the problem. The data structure
should be self-consistent regardless of how it is called.

## Consequences

### Positive
- Ghost counter inflation is fully blocked at two layers
- `set` maintains correct invariants in isolation — counters always match
  map contents
- Existing tests pass unchanged; 5 new regression tests added
- Minimal code change (no interface changes, no new types)

### Negative
- `set.Add()` is slightly more complex than the original due to the
  direction-change handling

### Test coverage added
- `TestSet_Add_DuplicateInbound`: same inbound peer added twice
- `TestSet_Add_DuplicateOutbound`: same outbound peer added twice
- `TestSet_Add_DirectionChange`: inbound peer replaced by outbound with same
  ID
- `TestSet_Remove_NonExistent`: remove on empty set (no underflow)
- `TestSet_Add_Remove_DuplicateCycle`: 10 duplicate adds + remove
