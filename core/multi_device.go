package core

import (
	"fmt"
	"strconv"
	"wa/signal/protocol"
	"wa/xmpp"

	"github.com/pkg/errors"
)

/*
// remove
{
  "Attrs": {
    "from": "2349159170075@s.whatsapp.net",
    "id": "2890645235",
    "t": "1640692855",
    "type": "devices"
  },
  "Children": [
    {
      "Attrs": {
        "device_hash": "2:0M1nDvxM"
      },
      "Children": [
        {
          "Attrs": {
            "jid": "2349159170075.0:3@s.whatsapp.net",
            "key-index": "1"
          },
          "Tag": "device"
        },
        {
          "Attrs": {
            "ts": "1640691730"
          },
          "Tag": "key-index-list"
        }
      ],
      "Tag": "remove"
    }
  ],
  "Tag": "notification"
}
*/

func New_Hook_MultiDeviceRemove(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		type_ := attrs[`type`]
		if type_ != `devices` {
			return nil
		}

		ch, ok := n.FindChildByTag(`remove`)
		if !ok {
			return nil
		}

		ch, ok = ch.FindChildByTag(`device`)
		if !ok {
			return nil
		}

		jid, ok := ch.GetAttr(`jid`)
		if !ok {
			return errors.New("no 'jid' attr")
		}

		recid, devid, e := split_jid(jid)
		if e != nil {
			return e
		}
		addr := protocol.NewSignalAddress(
			strconv.Itoa(int(recid)), devid)

		a.Store.DeleteIdentity(addr)
		a.Store.DeleteSession(addr)
		a.Store.DeleteSenderKey(addr)

		a.Store.DelMultiDevice(recid, devid)

		return nil
	}
}

/*
// add

	{
	  "Attrs": {
	    "from": "2349159170075@s.whatsapp.net",
	    "id": "1685660310",
	    "t": "1640692882",
	    "type": "devices"
	  },
	  "Children": [
	    {
	      "Attrs": {
	        "device_hash": "2:LjbWaW6U"
	      },
	      "Children": [
	        {
	          "Attrs": {
	            "jid": "2349159170075.0:4@s.whatsapp.net",
	            "key-index": "1"
	          },
	          "Tag": "device"
	        },
	        {
	          "Attrs": {
	            "ts": "1640692878"
	          },
	          "Data": "0a1208ddc1a4cc06108ef9ab8e0618012202000112407a28cfa688c13aeb3192063697797e275117d96a12f618e0ed8d196be09ceaabd649cf8892f1699ea53c1a53b2320edf81bbdc5b02857e211ea597a5e9e7be88",
	          "Tag": "key-index-list"
	        }
	      ],
	      "Tag": "add"
	    }
	  ],
	  "Tag": "notification"
	}
*/
func New_Hook_MultiDeviceAdd(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		type_ := attrs[`type`]
		if type_ != `devices` {
			return nil
		}

		ch, ok := n.FindChildByTag(`add`)
		if !ok {
			return nil
		}

		ch, ok = ch.FindChildByTag(`device`)
		if !ok {
			return nil
		}

		jid, ok := ch.GetAttr(`jid`)
		if !ok {
			return errors.New("no 'jid' attr")
		}

		recid, devid, e := split_jid(jid)
		if e != nil {
			return e
		}

		return a.Store.AddMultiDevice(recid, devid)
	}
}

/*
update:

	{
	  "Attrs": {
	    "from": "8613311112222@s.whatsapp.net",
	    "id": "3012531595",
	    "t": "1661436824",
	    "type": "devices"
	  },
	  "Children": [
	    {
	      "Attrs": {
	        "hash": "lWiW"
	      },
	      "Tag": "update"
	    }
	  ],
	  "Tag": "notification"
	}
*/
func New_Hook_MultiDeviceUpdate(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		type_ := attrs[`type`]
		if type_ != `devices` {
			return nil
		}

		_, ok := n.FindChildByTag(`update`)
		if !ok {
			return nil
		}

		from, ok := attrs[`from`]
		if !ok {
			return errors.New("no 'from' attr")
		}

		recid, _, e := split_jid(from)
		if e != nil {
			return e
		}

		// must use another goroutine, otherwise tcp would be blocking
		// usync() would freeze for 1 minute
		a.Wg.Add(1)
		go func() {
			defer a.Wg.Done()

			nr, e := a.usync_multi_device(
				fmt.Sprintf("%d@s.whatsapp.net", recid))
			if e != nil {
				return
			}

			a.handle_usync_multi_device_result(nr, recid)
		}()

		return nil
	}
}
