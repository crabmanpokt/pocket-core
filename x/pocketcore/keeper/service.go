package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/pokt-network/pocket-core/types"
	pc "github.com/pokt-network/pocket-core/x/pocketcore/types"
)

// "HandleRelay" - Handles an api (read/write) request to a non-native (external) blockchain
func (k Keeper) HandleRelay(ctx sdk.Ctx, relay pc.Relay) (*pc.RelayResponse, sdk.Error) {
	relayTimeStart := time.Now()
	// get the latest session block height because this relay will correspond with the latest session
	sessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	// get self node (your validator) from the current state
	selfAddrs := k.GetSelfAddress(ctx)
	// retrieve the nonNative blockchains your node is hosting
	hostedBlockchains := k.GetHostedBlockchains()
	// ensure the validity of the relay
	signer, maxPossibleRelays, err := relay.Validate(ctx, k.posKeeper, k.appKeeper, k, selfAddrs, hostedBlockchains, sessionBlockHeight)
	if err != nil {
		if pc.GlobalPocketConfig.RelayErrors {
			ctx.Logger().Error(
				fmt.Sprintf("could not validate relay for app: %s for chainID: %v with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					err.Error(),
				),
			)
			ctx.Logger().Debug(
				fmt.Sprintf(
					"could not validate relay for app: %s, for chainID %v on node(s) %s, at session height: %v, with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					selfAddrs.String(),
					sessionBlockHeight,
					err.Error(),
				),
			)
		}
		return nil, err
	}
	// store the proof before execution, because the proof corresponds to the previous relay
	relay.Proof.Store(maxPossibleRelays, signer)
	// attempt to execute
	respPayload, err := relay.Execute(hostedBlockchains)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("could not send relay with error: %s", err.Error()))
		return nil, err
	}
	// generate response object
	resp := &pc.RelayResponse{
		Response: respPayload,
		Proof:    relay.Proof,
	}
	// sign the response
	pk, err := k.GetSelfPrivKeyFromAddr(ctx, signer)
	if err != nil {
		return nil, err
	}
	sig, er := pk.Sign(resp.Hash())
	if er != nil {
		ctx.Logger().Error(
			fmt.Sprintf("could not sign response for address: %s with hash: %v, with error: %s",
				signer, resp.HashString(), er.Error()),
		)
		return nil, pc.NewKeybaseError(pc.ModuleName, er)
	}
	// attach the signature in hex to the response
	resp.Signature = hex.EncodeToString(sig)
	// track the relay time
	relayTime := time.Since(relayTimeStart)
	// add to metrics
	pc.GlobalServiceMetric().AddRelayTimingFor(relay.Proof.Blockchain, float64(relayTime.Milliseconds()))
	pc.GlobalServiceMetric().AddRelayFor(relay.Proof.Blockchain)
	return resp, nil
}

// "HandleChallenge" - Handles a client relay response challenge request
func (k Keeper) HandleChallenge(ctx sdk.Ctx, challenge pc.ChallengeProofInvalidData) (*pc.ChallengeResponse, sdk.Error) {
	// get self node (your validator) from the current state
	selfAddress := k.GetSelfAddress(ctx)
	sessionBlkHeight := k.GetLatestSessionBlockHeight(ctx)
	// get the session context
	sessionCtx, er := ctx.PrevCtx(sessionBlkHeight)
	if er != nil {
		return nil, sdk.ErrInternal(er.Error())
	}
	// get the application that staked on behalf of the client
	app, found := k.GetAppFromPublicKey(sessionCtx, challenge.MinorityResponse.Proof.Token.ApplicationPublicKey)
	if !found {
		return nil, pc.NewAppNotFoundError(pc.ModuleName)
	}
	// generate header
	header := pc.SessionHeader{
		ApplicationPubKey:  challenge.MinorityResponse.Proof.Token.ApplicationPublicKey,
		Chain:              challenge.MinorityResponse.Proof.Blockchain,
		SessionBlockHeight: sessionCtx.BlockHeight(),
	}
	// check cache
	session, found := pc.GetSession(header)
	// if not found generate the session
	if !found {
		var err sdk.Error
		blockHashBz, er := sessionCtx.BlockHash(k.Cdc, sessionCtx.BlockHeight())
		if er != nil {
			return nil, sdk.ErrInternal(er.Error())
		}
		session, err = pc.NewSession(sessionCtx, ctx, k.posKeeper, header, hex.EncodeToString(blockHashBz), int(k.SessionNodeCount(sessionCtx)))
		if err != nil {
			return nil, err
		}
		// add to cache
		pc.SetSession(session)
	}
	// validate the challenge
	err, selfNode := challenge.ValidateLocal(header, app.GetMaxRelays(), app.GetChains(), int(k.SessionNodeCount(sessionCtx)), session.SessionNodes, selfAddress)
	if err != nil {
		return nil, err
	}
	// store the challenge in memory
	challenge.Store(app.GetMaxRelays(), selfNode)
	// update metric
	pc.GlobalServiceMetric().AddChallengeFor(header.Chain)
	return &pc.ChallengeResponse{Response: fmt.Sprintf("successfully stored challenge proof for %s", challenge.MinorityResponse.Proof.ServicerPubKey)}, nil
}
