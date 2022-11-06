package core

import (
	"ahex"
	"wa/xmpp"
)

/*
{
  "Attrs": {
    "from": "s.whatsapp.net"
  },
  "Children": [
    {
      "Children": [
        {
          "Attrs": {
            "key": "AIzASyDR5yfaG7OG8sMTUj8kfQEb8T9pN8BM6Lk",
            "nonce": "ATcpqLIi05UFKrmxuYoopmiQsp5Y4wywTkLIPn0TgUZ0uDvVDa34VByseb9-kStEquAKtg9xbHlrOiZD3zkQ5XQb"
          },
          "Tag": "attestation"
        },
        {
          "Attrs": {
            "count": "10"
          },
          "Tag": "verify_apps"
        }
      ],
      "Tag": "safetynet"
    }
  ],
  "Tag": "ib"
}
*/

func New_Hook_Safetynet(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		if len(n.Children) == 0 {
			return nil
		}
		ch0 := n.Children[0]
		if ch0.Tag != `safetynet` {
			return nil
		}

		for _, ch := range ch0.Children {
			switch ch.Tag {
			case `attestation`:
				a.Store.SetSafetynetAttestation()
			case `verify_apps`:
				a.Store.SetSafetynetVerifyApps()
			}
		}
		return nil
	}
}

func (a *Acc) attestation(has_google_play bool) error {
	var code string
	var data []byte

	if has_google_play {
		code = `7`
		data = ahex.Dec(`373a20`) // "7: "
	} else {
		code = `1001`
		// my
		//data = ahex.Dec(`4174746573746174696f6e2041504920556e617661696c61626c652e20436f6e6e656374696f6e20726573756c7420636f64653a2039`)
		// emu
		data = ahex.Dec(`476f6f676c6520506c617920536572766963657320556e617661696c61626c652e20436f6e6e656374696f6e20726573756c7420636f64653a2039`)
	}

	e := a.Noise.WriteXmppNode(&xmpp.Node{
		Tag: `ib`,
		Children: []*xmpp.Node{
			{
				Tag: `attestation`,
				Children: []*xmpp.Node{
					{
						Tag: `error`,
						Attrs: []*xmpp.KeyValue{
							{Key: `code`, Value: code},
						},
						Data: data,
					},
				},
			},
		},
	})
	return e
}
func (a *Acc) verify_apps(has_google_play bool) error {
	var ch *xmpp.Node
	if has_google_play {
		ch = &xmpp.Node{
			Tag: `apps`,
			Attrs: []*xmpp.KeyValue{
				{Key: `actual_count`, Value: `0`},
			},
		}
	} else {
		ch = &xmpp.Node{
			Tag: `error`,
			Attrs: []*xmpp.KeyValue{
				{Key: `code`, Value: `1001`},
			},
			Data: ahex.Dec(`476f6f676c6520506c617920536572766963657320556e617661696c61626c652e20436f6e6e656374696f6e20726573756c7420636f64653a2039`),
		}
	}
	return a.Noise.WriteXmppNode(&xmpp.Node{
		Tag: `ib`,
		Children: []*xmpp.Node{
			{
				Tag: `verify_apps`,
				Children: []*xmpp.Node{
					ch,
				},
			},
		},
	})
}
