package def

const (
	// my
	Ev_AccCreate         = "acc_create"
	Ev_AccOn             = "acc_on"
	Ev_AccOff            = "acc_off"
	Ev_Connected         = "connected"
	Ev_NoiseDisconnected = "disconnected"
	Ev_Log               = "log"
	Ev_Noise_Location    = "noise_location"
	Ev_Heartbeat         = "heartbeat"
	Ev_Retry             = "retry"
	Ev_Success           = "success"
	Ev_Push              = "push"

	// dont modify string, it's Tag of xmpp
	Ev_call         = "call"
	Ev_iq           = "iq"
	Ev_message      = "message"
	Ev_receipt      = "receipt"
	Ev_ib           = "ib"
	Ev_notification = "notification"
)
