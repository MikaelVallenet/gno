# ADR: CLA Error Improvement — Structured Info + gnokey Hint

## Context

When CLA enforcement is enabled and a user tries to deploy a package without
signing, they get a terse error:

```
address g1xxx has not signed the required CLA
```

No guidance on what a CLA is, where to find the document, or how to sign it.
A user unfamiliar with the CLA realm is stuck.

The goal is to provide actionable feedback: explain what to do and give a
copy-pasteable gnokey command to sign the CLA.

## Decision

We split the responsibility across three layers, keeping each layer focused
on what it knows:

### 1. Keeper: structured error with CLA realm state

The keeper already detects CLA failures. Instead of returning a generic
`ErrUnauthorizedUser`, it now returns a `CLAUnsignedError` carrying metadata
read from the CLA realm: `RealmPath`, `Hash`, `URL`.

To read the realm's unexported variables (`requiredHash`, `claURL`), we
use the existing `queryEvalInternal` infrastructure. It sets `PkgPath` to
the CLA realm's package path on the GnoVM machine, which allows evaluating
identifiers in the realm's scope — including unexported variables. This is
the key mechanism: no exported getter functions are needed on the CLA realm.
`queryEvalInternal` uses a throwaway store, which is fine since the CLA
hash is always set by a prior committed govDAO proposal.

The `InfoKV()` output only carries CLA realm state — things the client
cannot know. Address, chain ID, and chain domain are client-side context
and are excluded from Info. (`Address` is kept in the struct for the
human-readable `Error()` message but not serialized into Info.)

Values in `InfoKV()` are sanitized (newlines stripped) to prevent injection
of arbitrary key-value pairs into the Info field.

### 2. Handler: populate ABCI Info field

The VM handler's `abciResult` detects `CLAUnsignedError` via
`errors.Cause()` and injects its key-value metadata into the ABCI `Info`
field:

```
cla.realm=gno.land/r/sys/cla
cla.hash=abc123hash
cla.url=https://github.com/gnolang/gno/blob/master/CLA.md
vm.version=0.1
```

The `Info` field is nondeterministic (per ABCI spec) and already used for
`vm.version=`. This makes the CLA data parseable by any client without
coupling the keeper to CLI-specific formatting.

### 3. gnokey: parse Info and format a user-friendly hint

A new `OnTxError` callback (mirroring the existing `OnTxSuccess`) in tm2
lets the gno.land `addpkg` command hook into broadcast failures. When the
Info field contains `cla.*` keys, gnokey formats a helpful message:

```
A Contributor License Agreement (CLA) must be signed before deploying packages.
It grants the necessary rights for your code to be used on-chain.
The CLA document is defined through a GovDAO governance proposal.

CLA document: https://github.com/gnolang/gno/blob/master/CLA.md

To sign the CLA, run:

  gnokey maketx call -pkgpath gno.land/r/sys/cla -func Sign -args abc123hash -gas-fee 100000ugnot -gas-wanted 2000000 -broadcast -chainid dev mykey
```

The command uses realm path and hash from the Info field, and address,
chain ID, and gas flags from gnokey's own context.

## Alternatives Considered

### Put everything in the keeper error message

The keeper could format the full hint (including the gnokey command) directly
in the error string. Rejected because:
- The keeper shouldn't know about gnokey CLI syntax.
- The error goes through `fmt.Sprintf("%#v", err)` in `ABCIResultFromError`,
  which mangles multi-line messages.
- Other clients (gnoweb, programmatic) would get gnokey-specific text.

### Add exported getter functions to the CLA realm

We could add `GetRequiredHash()` and `GetCLAURL()` to the realm and call
them via `callRealmString`. Rejected because:
- The keeper can already read realm state via expression evaluation in the
  package context — no new public API needed.
- Fewer changes to the CLA realm.

### Custom `evalRealmString` helper with live transaction store

We initially introduced a custom `evalRealmString` that uses the live
transaction store instead of `queryEvalInternal`. Rejected because:
- `queryEvalInternal` already exists and does the same thing.
- The CLA hash is always set by a prior committed govDAO proposal, so the
  throwaway store sees it fine.
- The custom helper introduced a new code path that could evaluate
  arbitrary expressions if misused.

### Generic `infoProvider` interface in the handler

Instead of type-asserting `CLAUnsignedError` directly, we considered a
generic `infoProvider` interface that any error could implement. Rejected
as over-engineering — `CLAUnsignedError` is currently the only error adding
to `res.Info`, and YAGNI applies.

## Consequences

- Users get clear, actionable feedback when CLA signing is required.
- The ABCI Info field becomes a channel for structured error metadata,
  parseable by any client.
- The `OnTxError` callback in tm2 is reusable for other error-specific
  hints in the future.
- The `CLAUnsignedError` type is registered with Amino, adding a new
  concrete type to the serialization registry.
- `queryEvalInternal` is called on the error path (tx already failing)
  with its own gas meter, so there is no impact on transaction gas.
