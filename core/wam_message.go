package core

import (
	"strconv"
	"time"

	"arand"
	"i"
	"wa/pb"
	"wa/wam"
)

func wam_media_type(media_t pb.Media_Type) int32 {
	var t int32 = 1
	switch media_t {
	case pb.Media_Text:
		t = 1
	case pb.Media_Image:
		t = 2
	case pb.Media_Contact:
		t = 7
	case pb.Media_Url:
		t = 9
	case pb.Media_Document:
		t = 8
	case pb.Media_Ptt:
		t = 5
	case pb.Media_Video:
		t = 3
	case pb.Media_Contact_Array:
		t = 13
	case pb.Media_Sticker:
		t = 16
	}
	return t
}

func (a *Acc) wam_message_send(
	media_t pb.Media_Type,
	send_begin time.Time,
	msg_type int,
	international bool,
) error {
	return a.AddWamEventBuf(wam.WamMessageSend, func(cc *wam.ClassChunk) {
		cc.Append(25, 1) // deviceSizeBucket
		cc.Append(23, 0) // e2eBackfill
		cc.Append(21, 0) // ephemeralityDuration
		cc.Append(22, 0) // isViewOnce

		if media_t == pb.Media_Image {
			cc.Append(8, 0) // mediaCaptionPresent
		}
		cc.Append(4, 0)                        // messageIsForward
		cc.Append(7, i.F(international, 1, 0)) // messageIsInternational
		cc.Append(3, wam_media_type(media_t))  // messageMediaType
		cc.Append(1, 1)                        // messageSendResult
		cc.Append(17, 0)                       // messageSendResultIsTerminal

		cc.Append(11, time.Since(send_begin).Milliseconds()) // messageSendT

		cc.Append(2, msg_type) // messageType

		if media_t == pb.Media_Sticker {
			cc.Append(18, 1) // stickerIsFirstParty
		}
		if media_t == pb.Media_Image { // image
			cc.Append(20, arand.Int(500, 1200)) // thumbSize
		}
	})
}

func (a *Acc) wam_e2e_message_send(
	media_t pb.Media_Type,
	e2eCiphertextType, e2eCiphertextVersion, e2eDestination int32,
) error {
	return a.AddWamEventBuf(wam.WamE2eMessageSend, func(cc *wam.ClassChunk) {
		cc.Append(5, e2eCiphertextType)
		cc.Append(6, e2eCiphertextVersion)
		cc.Append(4, e2eDestination)
		//cc.Append(2) // e2eFailureReason
		cc.Append(8, 1)          // e2eReceiverType
		cc.Append(1, 1)          // e2eSuccessful
		cc.Append(9, 0)          // encRetryCount
		if e2eDestination != 3 { // not sns
			cc.Append(7, wam_media_type(media_t)) // messageMediaType
		}
		cc.Append(3, 0) // retryCount
	})
}

func (a *Acc) wam_e2e_message_recv(
	media_t pb.Media_Type,
	timestamp int,
	e2eCiphertextType, e2eCiphertextVersion,
	e2eDestination int32,
) error {
	return a.AddWamEventBuf(wam.WamE2eMessageRecv, func(cc *wam.ClassChunk) {
		cc.Append(5, e2eCiphertextType)
		cc.Append(6, e2eCiphertextVersion)
		cc.Append(4, e2eDestination)
		cc.Append(8, 2) // e2eSenderType always 2
		cc.Append(1, 1) // e2eSuccessful

		cc.Append(7, wam_media_type(media_t)) // messageMediaType
		{
			is_offline := time.Now().Unix()-int64(timestamp) > 60
			cc.Append(9, i.F(is_offline, 1, 0)) // offline
		}
		cc.Append(3, 0) // retryCount
	})
}
func (a *Acc) wam_message_receive(
	media_t pb.Media_Type,
	timestamp int, // time > 1min == offline
	msg_type int, // 1:personal 2:group 3:broadcast 4:sns
	international bool,
) error {
	return a.AddWamEventBuf(wam.WamMessageReceive, func(cc *wam.ClassChunk) {
		cc.Append(9, 0)                        // isViewOnce
		cc.Append(4, i.F(international, 1, 0)) // messageIsInternational
		{
			is_offline := time.Now().Unix()-int64(timestamp) > 60
			cc.Append(5, i.F(is_offline, 1, 0)) // messageIsOffline
		}
		cc.Append(2, wam_media_type(media_t)) // messageMediaType
		cc.Append(6, arand.Int(-300, 300))    // messageReceiveT0
		cc.Append(7, arand.Int(15, 70))       // messageReceiveT1
		cc.Append(1, msg_type)                // messageType
	})
}

func wam_cipher_text_type(type_ string) int32 {
	switch type_ {
	case "msg":
		return 0
	case "pkmsg":
		return 1
	case "skmsg":
		return 2
	}
	return 0
}
func wam_cipher_text_ver(ver string) int32 {
	v, e := strconv.Atoi(ver)
	if e == nil {
		return int32(v)
	}
	return 2
}
