package core

import (
	"wa/xmpp"

	"go.mongodb.org/mongo-driver/bson"
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
          "Data": "080c0805",
          "Tag": "routing_info"
        }
      ],
      "Tag": "edge_routing"
    }
  ],
  "Tag": "ib"
}
*/

func New_Hook_RoutingInfo(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		// if contains child `edge_routing`
		ch, ok := n.FindChildByTag(`edge_routing`)
		if !ok {
			return nil
		}
		ch2, ok := ch.FindChildByTag(`routing_info`)
		if !ok {
			return nil
		}

		return a.Store.ModifyConfig(bson.M{
			`RoutingInfo`: ch2.Data,
		})
	}
}
