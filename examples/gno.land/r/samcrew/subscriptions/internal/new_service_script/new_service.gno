package main

import (
	"chain"
	"time"

	"gno.land/r/samcrew/subscriptions"
)

// Create a new subscription service
// To execute this function, you must use a MsgRun `gnokey maketx run`
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
// Example:
/*
gnokey maketx run \
  --gas-fee 1000000ugnot \
  --gas-wanted 10000000 \
  --broadcast \
  --chainid dev \
  --remote localhost:26657 \
  KEY \
  ./internal/new_service_script/new_service.gno
*/
func main() {
	// Create a new service with 30-day renewal period
	// Price: 1 GNOT per month
	subscriptions.NewService(
		cross,
		"Premium Plan",
		"Access to premium features with monthly billing",
		30*24*time.Hour,                                // 30 days
		chain.Coins{{Denom: "ugnot", Amount: 1000000}}, // 1 GNOT in native coins
		chain.Coins{},                                  // No GRC20 tokens
	)
}
