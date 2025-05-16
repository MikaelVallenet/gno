package gnolang

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

func TestConvertJSONValuePrimtive(t *testing.T) {
	cases := []struct {
		ValueRep string // Go representation
		Expected string // string representation
	}{
		// Boolean
		{"true", `{"T":"bool","V":true}`},
		{"false", `{"T":"bool","V":false}`},

		// int types
		{"int(42)", `{"T":"int","V":42}`}, // Needs to be quoted for amino
		{"int8(42)", `{"T":"int8","V":42}`},
		{"int16(42)", `{"T":"int16","V":42}`},
		{"int32(42)", `{"T":"int32","V":42}`},
		{"int64(42)", `{"T":"int64","V":42}`},

		// uint types
		{"uint(42)", `{"T":"uint","V":42}`},
		{"uint8(42)", `{"T":"uint8","V":42}`},
		{"uint16(42)", `{"T":"uint16","V":42}`},
		{"uint32(42)", `{"T":"uint32","V":42}`},
		{"uint64(42)", `{"T":"uint64","V":42}`},

		// Float types
		{"float32(3.14)", `{"T":"float32","V":3.14}`},
		{"float64(3.14)", `{"T":"float64","V":3.14}`},

		// String type
		{`"hello world"`, `{"T":"string","V":"hello world"}`},

		// UntypedRuneType
		{`'A'`, `{"T":"int32","V":65}`},

		// DataByteType (assuming DataByte is an alias for uint8)
		{"uint8(42)", `{"T":"uint8","V":42}`},

		// Byte slice
		{`[]byte("AB")`, `{"T":"[]uint8","V":{"@type":"/gno.SliceValue","Base":{"@type":"/gno.RefValue"},"Offset":"0","Length":"2","Maxcap":"8"}}`},

		// Byte array
		{`[2]byte{0x41, 0x42}`, `{"T":"[2]uint8","V":{"@type":"/gno.RefValue"}}`},

		// XXX: BigInt
		// XXX: BigDec
	}

	for _, tc := range cases {
		t.Run(tc.ValueRep, func(t *testing.T) {
			m := NewMachine("testdata", nil)
			defer m.Release()

			nn := MustParseFile("testdata.gno",
				fmt.Sprintf(`package testdata; var Value = %s`, tc.ValueRep))
			m.RunFiles(nn)
			m.RunDeclaration(ImportD("testdata", "testdata"))

			tps := m.Eval(Sel(Nx("testdata"), "Value"))
			require.Len(t, tps, 1)

			tv := tps[0]

			rep, err := JSONExportTypedValue(tv)
			require.NoError(t, err)

			require.Equal(t, string(tc.Expected), string(rep))
		})
	}
}

func TestConvertJSONValueStruct(t *testing.T) {
	const StructsFile = `
package testdata

// E struct, impement error
type E struct { S string }

func (e *E) Error() string { return e.S }
`
	t.Run("null pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const expected = "null"

		nn := MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = MustParseFile("testdata.gno", `package testdata; var Value *E = nil`)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("without pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		const expected = `{"$error":"Hello World"}`

		nn := MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		tt, vv := amino.MustMarshalJSONAny(tv.T), amino.MustMarshalJSONAny(tv.V)
		fmt.Printf("before T: (%s) - V: (%s)\n", string(tt), string(vv))

		rep, err := JSONExportTypedValue(tv)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("with pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		const expected = `{"$error":"Hello World"}`

		nn := MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = &E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})
}

func TestConvertJSONValueRecusiveStruct(t *testing.T) {
	const RecursiveValueFile = `
package testdata
type Recursive struct {
	Nested *Recursive
}
var RecursiveStruct = &Recursive{}
func init() {
	RecursiveStruct.Nested = RecursiveStruct
}
`
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := MustParseFile("testdata.gno", RecursiveValueFile)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "RecursiveStruct"))
	require.Len(t, tps, 1)
	tv := tps[0]

	val := ExportValue(tv)
	ret := amino.MustMarshalJSONAny(val)
	fmt.Printf("rec: %s\n", string(ret))

	fmt.Printf("VAL %#v\n", val)
	ref := val.V.(PointerValue).Base.(RefValue)
	require.False(t, ref.ObjectID.IsZero())

	data, err := JSONExportTypedValue(tv)
	require.NoError(t, err)
	fmt.Println(string(data))
}
