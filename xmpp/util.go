package xmpp

import "fmt"

const NotJid = 0
const NormalJid = 1
const DevJid = 2

func analyze_jid(jid string) (user int, agent, device uint8, server string, jid_type int) {
	_, e := fmt.Sscanf(jid,
		"%d.%d:%d@%s", &user, &agent, &device, &server)

	if e == nil {
		jid_type = DevJid
		return
	}

	_, e = fmt.Sscanf(jid,
		"%d@%s", &user, &server)
	if e == nil {
		jid_type = NormalJid
		return
	}
	return
}

