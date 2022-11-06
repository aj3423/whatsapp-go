package core

import (
	"fmt"

	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/keys/prekey"
	"wa/signal/protocol"
	"wa/signal/session"
	"wa/signal/util/bytehelper"
	"wa/signal/util/optional"

	"github.com/pkg/errors"
)

/*
This function doesn't expand multi-device,
param `jid_arr` should be already expanded.
*/
func (a *Acc) ensure_session_builder(
	jid_arr []string,
) (map[string]*session.Builder, error) {

	ret := map[string]*session.Builder{}

	no_session_jids := []string{}

	for _, jid := range jid_arr {
		recid, devid, e := split_jid(jid)
		if e != nil {
			return nil, e
		}

		peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)

		if a.Store.ContainsSession(peer_addr) {
			sb := session.NewBuilder(
				a.Store, a.Store, a.Store, a.Store, peer_addr)
			ret[jid] = sb

			continue
		}
		// if no previous session
		no_session_jids = append(no_session_jids, jid)
	}

	// if all session presents
	if len(no_session_jids) == 0 {
		return ret, nil
	}

	// 1. get_encrypt
	emap, e := a.get_encrypt(no_session_jids, `key`)
	if e != nil {
		return nil, errors.Wrap(e, "get_encrypt Fail")
	}

	for _, jid := range no_session_jids {
		user, ok := emap[jid]
		if !ok {
			return nil, errors.Wrap(e, `WTF, missing Encrypt Result for jid `+jid)
		}

		recid, devid, _ := split_jid(jid)

		// if bot account has no prekey left
		var prekey_ ecc.ECPublicKeyable = nil
		var prekey_id *optional.Uint32 = optional.NewEmptyUint32()
		if user.Prekey != nil {
			prekey_ = ecc.NewDjbECPublicKey(user.Prekey)
			prekey_id = optional.NewOptionalUint32(uint32(user.PrekeyId))
		} else {
			a.Log.Debug("prekey is nil")
		}

		// 2. create session
		pkb := prekey.NewBundle(
			user.RegId,
			devid,
			prekey_id,
			user.SpkId,
			prekey_,
			ecc.NewDjbECPublicKey(user.Spk),
			bytehelper.SliceToArray64(user.SpkSig),
			identity.NewKeyFromBytes(bytehelper.SliceToArray(user.Identity)),
		)

		peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)
		sb := session.NewBuilder(
			a.Store, a.Store, a.Store, a.Store, peer_addr)
		if e := sb.ProcessBundle(pkb); e != nil {
			return nil, errors.Wrap(e, `fail sb.ProcessBundle `)
		}
		ret[jid] = sb
	}

	return ret, nil
}
