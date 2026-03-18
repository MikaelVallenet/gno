package banker

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func BenchmarkCompactCoins(b *testing.B) {
	counts := []int{1, 5, 10, 50}
	for _, n := range counts {
		denoms := make([]string, n)
		amounts := make([]int64, n)
		for i := 0; i < n; i++ {
			denoms[i] = fmt.Sprintf("denom%d", i)
			amounts[i] = int64(1000 * (i + 1))
		}
		b.Run(fmt.Sprintf("coins=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				CompactCoins(denoms, amounts)
			}
		})
	}
}

func BenchmarkAddressParsing(b *testing.B) {
	addr := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		crypto.MustAddressFromString(addr)
	}
}

func BenchmarkBankerSendOverhead(b *testing.B) {
	// Benchmarks the full Go-side overhead of a SendCoins call
	// (address parsing + coin compaction), excluding store I/O.
	from := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	to := "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
	counts := []int{1, 5, 10, 50}
	for _, n := range counts {
		denoms := make([]string, n)
		amounts := make([]int64, n)
		for i := 0; i < n; i++ {
			denoms[i] = fmt.Sprintf("denom%d", i)
			amounts[i] = int64(1000 * (i + 1))
		}
		b.Run(fmt.Sprintf("coins=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				crypto.MustAddressFromString(from)
				crypto.MustAddressFromString(to)
				CompactCoins(denoms, amounts)
			}
		})
	}
}
