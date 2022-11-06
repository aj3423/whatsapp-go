package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"

	"afs"
	"ajson"
	"algo"
	"github.com/beevik/etree"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"wa/def"
	"wa/def/clone"
)

var dir string
var out string

func init() {
	flag.StringVar(&dir, "dir", "./com.whatsapp", "")
	flag.StringVar(&out, "out", "./clone.json", "")
}

func main() {

	flag.Parse()

	gene := &clone.Gene{}

	{
		doc := etree.NewDocument()
		if e := doc.ReadFromFile(dir + "/shared_prefs/keystore.xml"); e == nil {
			client_static_keypair_pwd_enc := doc.FindElements("//map/string[@name='client_static_keypair_pwd_enc']")[0].Text()
			j, e := ajson.Parse(client_static_keypair_pwd_enc)
			if e != nil {
				panic(e)
			}

			staticI_enc, e1 := algo.B64PadDec(j.GetIndex(1).String())

			iv, e2 := algo.B64PadDec(j.GetIndex(2).String())

			salt, e3 := algo.B64PadDec(j.GetIndex(3).String())

			password := j.GetIndex(4).String()

			if e1 != nil || e2 != nil || e3 != nil {
				panic(errors.New(`fail B64PadDec`))
			}
			key := algo.PbkdfSha1(
				append(def.RC2_FIXED_25, []byte(password)...),
				salt, 16, 16,
			)

			bs, e := algo.AesOfbDecrypt(staticI_enc, key, iv, &algo.None{})
			if e != nil {
				panic(e)
			}
			gene.StaticPriv = bs[0:0x20]
			gene.StaticPub = bs[0x20:]
		}

	}
	{
		doc := etree.NewDocument()
		if e := doc.ReadFromFile(dir + "/shared_prefs/com.whatsapp_preferences_light.xml"); e == nil {
			x := doc.FindElements("//map/string[@name='push_name']")
			if len(x) > 0 {
				gene.Nick = x[0].Text()
			}

			gene.Fdid = doc.FindElements("//map/string[@name='phoneid_id']")[0].Text()

			z := doc.FindElements("//map/string[@name='routing_info']")
			if len(z) > 0 {
				ri := z[0].Text()
				var e error
				gene.RoutingInfo, e = base64.RawURLEncoding.DecodeString(ri)
				if e != nil {
					panic(e)
				}
			}

			y := doc.FindElements("//map/string[@name='last_datacenter']")
			if len(y) > 0 {
				gene.NoiseLocation = y[0].Text()
			}
		}
	}
	{
		doc := etree.NewDocument()
		if e := doc.ReadFromFile(dir + "/shared_prefs/ab-props.xml"); e == nil {
			x := doc.FindElements("//map/string[@name='ab_props:sys:config_key']")
			if len(x) > 0 {
				gene.AbPropsConfigKey = x[0].Text()
			}

			y := doc.FindElements("//map/string[@name='ab_props:sys:config_hash']")
			if len(y) > 0 {
				gene.AbPropsConfigHash = y[0].Text()
			}
		}
	}
	{
		doc := etree.NewDocument()
		if e := doc.ReadFromFile(dir + "/shared_prefs/com.whatsapp_preferences.xml"); e == nil {
			x := doc.FindElements("//map/string[@name='server_props:config_key']")
			if len(x) > 0 {
				gene.ServerPropsConfigKey = x[0].Text()
			}

			y := doc.FindElements("//map/string[@name='server_props:config_hash']")
			if len(y) > 0 {
				gene.ServerPropsConfigHash = y[0].Text()
			}
		}
	}
	db, e := gorm.Open(sqlite.Open(dir+"/databases/axolotl.db"), &gorm.Config{})
	if e != nil {
		panic(e)
	}
	{
		{
			iden := &clone.Identity{RecipientId: -1}
			if e = db.First(iden).Error; e != nil {
				panic(e)
			}
			gene.Identity = *iden
			// djb key
			if len(gene.Identity.PublicKey) == 33 && gene.Identity.PublicKey[0] == 5 {
				gene.Identity.PublicKey = gene.Identity.PublicKey[1:]
			}
		}
		{
			prekeys := []clone.Prekey{}
			if e = db.Find(&prekeys).Error; e != nil {
				panic(e)
			}
			gene.Prekeys = prekeys
		}
	}
	{
		{
			spk := &clone.SignedPrekey{}
			if e = db.First(spk).Error; e != nil {
				panic(e)
			}
			gene.SignedPrekey = *spk
		}
	}

	bs, _ := json.MarshalIndent(gene, "", "  ")
	afs.Write(out, bs)
}
