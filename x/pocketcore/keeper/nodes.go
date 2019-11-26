package keeper

import (
	"github.com/pokt-network/pocket-core/types"
	"github.com/pokt-network/pocket-core/x/nodes/exported"
	sdk "github.com/pokt-network/posmint/types"
)

func (k Keeper) GetAllNodes(ctx sdk.Context) []exported.ValidatorI {
	return k.posKeeper.GetAllValidators(ctx)
}

func (k Keeper) GetNodeFromPublicKey(ctx sdk.Context, pubKey string) (node exported.ValidatorI, found bool) {
	// get the node at the session context
	nodeAddr, err := k.AddressFromPubKeyString(pubKey)
	if err != nil {
		return nil, false
	}
	return k.GetNode(ctx, nodeAddr)
}

func (k Keeper) GetNode(ctx sdk.Context, address sdk.ValAddress) (node exported.ValidatorI, found bool) {
	return k.posKeeper.GetValidator(ctx, address)
}

func (k Keeper) GetNodeChains(ctx sdk.Context, address sdk.ValAddress) (chains map[string]struct{}, found bool) {
	node, found := k.posKeeper.GetValidator(ctx, address)
	if !found {
		return nil, false
	}
	return node.GetChains(), true
}

func (k Keeper) GetNodeServiceURL(ctx sdk.Context, address sdk.ValAddress) (serviceURL string, found bool) {
	node, found := k.posKeeper.GetValidator(ctx, address)
	if !found {
		return "", false
	}
	return node.GetServiceURL(), true
}

// todo create store in pos module for efficiency
func (k Keeper) GetAllNodesForChain(ctx sdk.Context, chain string) (nodes []exported.ValidatorI) {
	return nil
}

func (k Keeper) GetSelfNode(ctx sdk.Context) (node exported.ValidatorI) {
	return nil // todo
}

func (k Keeper) GetHostedBlockchains(ctx sdk.Context) types.HostedBlockchains {
	return types.HostedBlockchains{} // todo
}

func (k Keeper) AwardCoinsForRelays(ctx sdk.Context, relays int64, toAddr sdk.ValAddress) {
	k.posKeeper.AwardCoinsTo(ctx, sdk.NewInt(relays), toAddr)
}