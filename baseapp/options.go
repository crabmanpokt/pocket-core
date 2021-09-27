// nolint: golint
package baseapp

import (
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/pokt-network/pocket-core/types"
)

func (app *BaseApp) SetName(name string) {
	app.name = name
}

// SetAppVersion sets the application's version string.
func (app *BaseApp) SetAppVersion(v string) {
	app.appVersion = v
}

func (app *BaseApp) SetDB(db dbm.DB) {
	app.db = db
}

func (app *BaseApp) SetInitChainer(initChainer sdk.InitChainer) {
	app.initChainer = initChainer
}

func (app *BaseApp) SetBeginBlocker(beginBlocker sdk.BeginBlocker) {
	app.beginBlocker = beginBlocker
}

func (app *BaseApp) SetEndBlocker(endBlocker sdk.EndBlocker) {
	app.endBlocker = endBlocker
}

func (app *BaseApp) SetAnteHandler(ah sdk.AnteHandler) {
	app.anteHandler = ah
}

func (app *BaseApp) SetAddrPeerFilter(pf sdk.PeerFilter) {
	app.addrPeerFilter = pf
}

func (app *BaseApp) SetIDPeerFilter(pf sdk.PeerFilter) {
	app.idPeerFilter = pf
}

func (app *BaseApp) SetFauxMerkleMode() {
	app.fauxMerkleMode = true
}
