package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper of the mint store
type Keeper struct {
	cdc              *codec.Codec
	storeKey         sdk.StoreKey
	paramSpace       params.Subspace
	sk               types.StakingKeeper
	supplyKeeper     types.SupplyKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
	sk types.StakingKeeper, supplyKeeper types.SupplyKeeper, feeCollectorName string,
) Keeper {

	// ensure mint module account is set
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace.WithKeyTable(types.ParamKeyTable()),
		sk:               sk,
		supplyKeeper:     supplyKeeper,
		feeCollectorName: feeCollectorName,
	}
}

//______________________________________________________________________

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(types.MinterKey, b)
}

//______________________________________________________________________

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

//______________________________________________________________________

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.sk.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.sk.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.supplyKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

// CalculateCoin calculate coins per month with fixed total
// CalculateCoin to be used in BeginBlocker.
func (k Keeper) CalculateCoin(ctx sdk.Context, params types.Params) sdk.Int {
	var (
		unitCoin = params.UnitCoin
		count    = int64(0)
		cycle    = ctx.BlockHeight() / params.BlocksPerUnit
	)
	for {
		if count >= cycle {
			break
		}
		unitCoin = unitCoin.Mul(params.Decrease).QuoRaw(100) //Get rewards in the current period
		count++
	}

	if unitCoin.LTE(sdk.ZeroInt()) {
		return sdk.ZeroInt()
	}
	nowTotalSupply := k.GetNowTotalSupply(ctx)
	if nowTotalSupply.LT(sdk.ZeroInt()) {
		k.SetNowTotalSupply(ctx, params.TotalSupply.Sub(unitCoin))
	} else if nowTotalSupply.GT(sdk.ZeroInt()) && nowTotalSupply.GTE(unitCoin) {
		k.SetNowTotalSupply(ctx, nowTotalSupply.Sub(unitCoin))
	} else if nowTotalSupply.GT(sdk.ZeroInt()) && nowTotalSupply.LT(unitCoin) {
		k.SetNowTotalSupply(ctx, nowTotalSupply.Sub(nowTotalSupply))
		unitCoin = nowTotalSupply
	} else {
		unitCoin = sdk.ZeroInt()
	}
	//fmt.Println("height:", ctx.BlockHeight(), "nowCycle:", cycle, "newCoin:", unitCoin, "nowTotalSupply", k.GetNowTotalSupply(ctx))
	return unitCoin
}

func (k Keeper) GetNowTotalSupply(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(types.NowTotalSupply)) {
		params := k.GetParams(ctx)
		store.Set([]byte(types.NowTotalSupply), k.cdc.MustMarshalBinaryBare(params.TotalSupply))
	}
	var data sdk.Int
	bz := store.Get([]byte(types.NowTotalSupply))
	k.cdc.MustUnmarshalBinaryBare(bz, &data)
	return data
}

func (k Keeper) SetNowTotalSupply(ctx sdk.Context, supply sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.NowTotalSupply), k.cdc.MustMarshalBinaryBare(supply))
}