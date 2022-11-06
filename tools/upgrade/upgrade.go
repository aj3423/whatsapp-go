package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"os"
	"strings"

	"afs"
	"ahex"
	"algo"
	"aregex"
	"run"
	"wa/def"

	"github.com/fatih/color"
)

func check_change(tag string, prev, curr []byte) {
	if bytes.Equal(prev, curr) {
		color.HiGreen("%s: same", tag)
	} else {
		color.HiYellow("%s: changed", tag)
		color.HiMagenta("prev:\n" + hex.Dump(prev))
		color.HiBlue("curr:\n" + hex.Dump(curr))
	}
}

var biz bool

func init() {
	flag.BoolVar(&biz, "biz", false, "")
}

var DIR string = `/root/Downloads/up/personal`
var __pkg_name__ = `com.whatsapp`
var __apk_fn__ = `wa`

func main() {
	flag.Parse()

	if biz {
		__pkg_name__ = `com.whatsapp.w4b`
		__apk_fn__ = `wa_biz`
		DIR = `/root/Downloads/up/biz`
	}

	os.Chdir(DIR)
	if !afs.Exist(__pkg_name__) {
		color.HiRed("no dir: %s/%s", DIR, __pkg_name__)
		color.HiYellow("pull from '/data/data/%s' first", __pkg_name__)
		os.Exit(1)
	}
	if !afs.Exist(__apk_fn__) {
		color.HiRed("no dir: %s/%s", DIR, __apk_fn__)
		color.HiYellow("call 'apktool d " + __apk_fn__ + "' first")
		os.Exit(1)
	}

	{ // VERSION
		out, err, ec := run.RunCommand(DIR, `aapt`, `d`, `badging`, __apk_fn__+`.apk`)
		if ec != 0 {
			panic(`fail ag: ` + err)
		}
		s := aregex.Search(out, "versionName='(.+?)'")
		color.HiYellow(`VERSION: %s`, s)
	}
	{ // CLASSES_MD5
		if !afs.Exist(`classes.dex`) {
			color.HiRed("no classes.dex")
			os.Exit(1)
		}
		bs := afs.Read(`classes.dex`)
		color.HiYellow(`CLASSES_MD5: %s`, algo.Md5Str(bs))
	}
	{ // SVR_PUB
		bs := afs.Read(__pkg_name__ + `/files/decompressed/libs.spk.zst/libwhatsapp.so`)
		p := bytes.Index(bs, def.SVR_PUB)
		if p <= 0 {
			color.HiYellow("%s: changed", "SVR_PUB")
			os.Exit(1)
		} else {
			color.HiGreen("%s: same", "SVR_PUB")
		}
	}
	{ // SALT
		// 1. find file contains string `/res/drawable....`
		out, err, ec := run.RunCommand(DIR, `ag`, `-l`, `--no-break`, `/res/drawable-hdpi/about_logo.png`, __apk_fn__)
		if ec != 0 {
			panic(`fail ag: ` + err)
		}
		fn := strings.Trim(out, "\n")
		// 2. regex search
		s := aregex.Search(afs.ReadStr(fn), "const-string v1, \"([^\"]+?)\"")
		curr, _ := algo.B64Dec(s)
		check_change(`SALT`, def.SALT, curr)
	}
	{ // ABOUT_LOGO
		curr := afs.Read(__apk_fn__ + `/res/drawable-hdpi/about_logo.png`)
		check_change(`ABOUT_LOGO`, def.ABOUT_LOGO(biz), curr)
	}
	{ // SignatureHash
		out, err, ec := run.RunCommand(DIR, `keytool`, `-printcert`, `-jarfile`, __apk_fn__+`.apk`)
		if ec != 0 {
			panic(`fail ag: ` + err)
		}
		// 2. regex search
		s := aregex.Search(out, "SHA256: (.+?)\n")
		s = strings.ReplaceAll(s, ":", "")
		curr := ahex.Dec(s)

		prev, _ := algo.B64Dec(def.SignatureHash)
		check_change(`SignatureHash`, prev, curr)
	}
	{ // SIGNATURE
		fn := __apk_fn__ + `/original/META-INF/WHATSAPP.DSA`
		out, err, ec := run.RunCommand(DIR, `openssl`, `pkcs7`, `-inform`, `DER`, `-in`, fn, `-print_certs`, `-text`)
		if ec != 0 {
			panic(`fail ag: ` + err)
		}
		s := aregex.Search(out, "-----BEGIN CERTIFICATE-----([^-]+?)-----END CERTIFICATE-----")
		s = strings.ReplaceAll(s, "\r", "")
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\t", "")
		s = strings.ReplaceAll(s, " ", "")
		curr, _ := algo.B64Dec(s)
		check_change(`SIGNATURE`, def.SIGNATURE, curr)
	}
	{ // RC2_FIXED_25
		// 1. find file contains string
		out, err, ec := run.RunCommand(DIR, `ag`, `-l`, `--no-break`, "-Q", `A\u0004\u001d@\u0011\u0018V\u0091\u0002\u0090\u0088\u009f\u009eT(3{;ES`, __apk_fn__)
		if ec != 0 {
			panic(`fail ag: ` + err)
		}
		if strings.Contains(out, `.smali`) {
			check_change(`RC2_FIXED_25`, []byte{}, []byte{})
		} else {
			check_change(`RC2_FIXED_25`, []byte{}, []byte(`str not found`))
		}
	}
	{
		//TODO
		color.Blue("Now manually verify WA_41 ED_..")
		color.Blue("Now manually verify dict_0/1/2 in xmpp/def.go")
		color.Blue("Now manually verify WAM.header [57 41 4d 05] in wam/wam.go")
	}
}
