package crypto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"ahex"
	"algo"
	"arand"
	"wa/def"
	"wa/signal/ecc"

	"golang.org/x/crypto/curve25519"
)

func NewECKeyPair() ([]byte, []byte) {
	kp, _ := ecc.GenerateKeyPair()

	pub := kp.PublicKey().PublicKey()
	priv := kp.PrivateKey().Serialize()

	return priv[:], pub[:]
}

func Curve25519Agree(priv_me, pub_svr []byte) []byte {
	var priv [32]byte
	copy(priv[:], priv_me)
	var pub [32]byte
	copy(pub[:], pub_svr)

	var result [32]byte
	curve25519.ScalarMult(&result, &priv, &pub)
	return result[:]
}

func CalcToken(phone string, is_biz bool) []byte {
	password := []byte{}
	password = append(password, def.PKG_NAME(is_biz)...)
	password = append(password, def.ABOUT_LOGO(is_biz)...)
	derived := algo.PbkdfSha1(password, def.SALT, 0x40, 128)

	x := []byte{}
	x = append(x, def.SIGNATURE...)
	x = append(x, def.AppCodeHash(is_biz)...)
	x = append(x, []byte(phone)...)
	r := algo.HmacSha1(x, derived)

	return r
}

var Param_Order = []string{
	"reason",
	"method",
	"read_phone_permission_granted",
	"lc",
	"offline_ab",
	"in",
	"backup_token",
	"lg",
	"e_regid",
	"mistyped",
	"id",
	"authkey",
	"e_skey_sig",
	"hasav",
	"action_taken",
	"token",
	"expid",
	"e_ident",
	"previous_screen",
	"rc",
	"sim_mcc",
	"simnum",
	"entered",
	"sim_state",
	"client_metrics",
	"cc",
	"e_skey_id",
	"mnc",
	"sim_mnc",
	"fdid",
	"funnel_id",
	"e_skey_val",
	"hasinrc",
	"network_radio_type",
	"mcc",
	"network_operator_name",
	"sim_operator_name",
	"e_keytype",
	"pid",
	"code",
	"current_screen",
	"vname",
}

func BuildUrl(param map[string]string) (string, error) {
	lst := []string{}
	for _, odr_key := range Param_Order {
		if v, ok := param[odr_key]; ok {
			lst = append(lst, odr_key+`=`+algo.UrlEnc(v))
		}
	}
	if len(lst) != len(param) {
		return "", errors.New("some url param no defined in BuildUrl" + fmt.Sprintf("%v", lst) + ", " + fmt.Sprintf("%v", param))
	}
	return strings.Join(lst, "&"), nil
}

func BE2U32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// U32 To BigEndian
func U322BE(num uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:], num)
	return buf
}

func LE2U8(data []byte) uint8 {
	return data[0]
}
func U82LE(num uint8) []byte {
	return []byte{num}
}

func LE2U16(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data)
}
func U162LE(num uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf[0:], num)
	return buf
}

func LE2U32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}
func U322LE(num uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf[0:], num)
	return buf
}

func LE2U64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}
func U642LE(num uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf[0:], num)
	return buf
}

// U32 To BigEndian (24 bit)
func U322BE_24(num uint32) []byte {
	return U322BE(num)[1:]
}

// 3 bytes -> uint32
func BE2U24(data []byte) uint32 {
	b := append([]byte{0}, data...)
	return binary.BigEndian.Uint32(b)
}

func RandomPadMsg(data []byte) []byte {
	{
		//color.HiMagenta(hex.Dump(data))
	}

	padded := make([]byte, len(data))
	copy(padded, data)

	pad_len := arand.Int(1, 0x10) // 1~15

	for i := 0; i < pad_len; i++ {
		padded = append(padded, byte(pad_len))
	}
	return padded
}
func UnPadMsg(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, errors.New(`invalid data to unpad: ` + ahex.Enc(data))
	}
	last := int(data[len(data)-1])

	p := len(data) - 1
	for i := 0; i < last; i++ {
		if int(data[p]) != last {
			return nil, errors.New(`invalid data to unpad: ` + ahex.Enc(data))
		}
		p--
	}
	if len(data) < last {
		return nil, errors.New(`invalid data to unpad: ` + ahex.Enc(data))
	}
	return data[0 : len(data)-last], nil
}
