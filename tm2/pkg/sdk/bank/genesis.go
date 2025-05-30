package bank

import (
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// GenesisState - all state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) GenesisState {
	return GenesisState{params}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}

// InitGenesis - Init store state from genesis data
func (bank BankKeeper) InitGenesis(ctx sdk.Context, data GenesisState) {
	if err := ValidateGenesis(data); err != nil {
		panic(err)
	}

	if err := bank.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func (bank BankKeeper) ExportGenesis(ctx sdk.Context) GenesisState {
	params := bank.GetParams(ctx)

	return NewGenesisState(params)
}
