# PR5269: Constant Conversion Overflow Checks

## Context

Large `uint64` constants (e.g., `math.MaxUint64`) were bypassing overflow validation
when converting to smaller types. The root cause was that `validate` comparisons
cast `uint64` values to `int64` first, which silently wrapped large values and
made the bounds check pass incorrectly.

## Decision

### Fix: Remove `int64()` cast in uint64 overflow comparisons

In `gnovm/pkg/gnolang/values_conversions.go`, the `Uint64Kind` conversion cases
compared `int64(tv.GetUint64()) <= math.MaxInt8` etc. Casting a large uint64 to
int64 wraps it to a negative number, which is always `<= MaxInt8`. The fix removes
the `int64()` cast so the comparison stays in uint64 space.

### Error messages: match Go's logic

Go's type checker (`src/go/types/conversions.go`) uses two distinct error formats:

```go
if isInteger(x.typ()) && isInteger(u) {
    cause = "constant %s overflows %s"
} else {
    cause = "cannot convert %s to type %s"
}
```

- **Both source and target are integers** → `"constant VALUE overflows TARGET"`
  (e.g., `constant 18446744073709551615 overflows Int8Kind`)
- **Otherwise** (e.g., float→int) → `"cannot convert constant of type FROM to TO"`
  (e.g., `cannot convert constant of type Float32Kind to Int32Kind`)

The `validate` closure in `ConvertTo` mirrors this with an `isIntegerKind` helper.

### int→string const conversions are valid

`string(typed_int_const)` is valid Go (produces a rune string, e.g., `string(int8(65))` → `"A"`).
All `validate(XxxKind, StringKind, nil)` calls were removed as they would reject valid Go code.

Note: the preprocessing path in `preprocess.go:1539` has an `isIntNum(ct)` guard that
skips `ConvertTo` with `isConst=true` when the target is `StringType`. So even if
validate calls existed, they would be dead code through the normal preprocessing path.
The actual int→string conversion happens at runtime via `op_expressions.go:727`.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/values_conversions.go` | `ConvertTo` function with `validate` closure |
| `gnovm/pkg/gnolang/preprocess.go:1539` | `isIntNum(ct)` guard — controls which const conversions go through `ConvertTo` |
| `gnovm/pkg/gnolang/op_expressions.go:727` | Runtime conversion path (`isConst=false`) |
| `gnovm/tests/files/convert9*.gno` | Filetests for uint64 overflow and int→string |

## Consequences

- `const a uint64 = math.MaxUint64; int8(a)` now correctly errors at preprocess time
- Float→int const errors use "cannot convert" (not "overflows"), matching Go
- `string(const_typed_int)` works correctly, matching Go behavior
