package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"strconv"
	"strings"

	"ajson"
	"algo"
	"wa/def"
	"wa/pb"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func MediaTypeStr(t pb.Media_Type) string {
	return strings.ToLower(t.String())
}
func MediaTypeInt(t string) pb.Media_Type {
	switch strings.ToLower(t) {
	case `text`:
		return pb.Media_Text
	case `url`:
		return pb.Media_Url
	case `document`:
		return pb.Media_Document
	case `image`:
		return pb.Media_Image
	case `ptt`:
		return pb.Media_Ptt
	case `sticker`:
		return pb.Media_Sticker
	case `video`:
		return pb.Media_Video
	case `contact`:
		return pb.Media_Contact
	case `contact_array`:
		return pb.Media_Contact_Array
	}
	return pb.Media_Unknown
}
func NewMedia(
	media_t pb.Media_Type,
) (Media, error) {
	var ret Media
	switch media_t {
	case pb.Media_Text:
		ret = &Text{}
	case pb.Media_Url:
		ret = &Url{P: &pb.Url{}}
	case pb.Media_Document:
		ret = &Document{P: &pb.Document{}}
	case pb.Media_Sticker:
		ret = &Sticker{P: &pb.Sticker{}}
	case pb.Media_Ptt:
		ret = &Ptt{P: &pb.Ptt{}}
	case pb.Media_Image:
		ret = &Image{P: &pb.Image{}}
	case pb.Media_Video:
		ret = &Video{P: &pb.Video{}}
	case pb.Media_Contact:
		ret = &Contact{P: &pb.Contact{}}
	case pb.Media_Contact_Array:
		ret = &ContactArray{P: &pb.ContactArray{}}
	default:
		return nil, errors.New(`unsupported media type: ` + strconv.Itoa(int(media_t)))
	}
	return ret, nil
}
func NewMediaFromBytes(
	media_t pb.Media_Type,
	pb_data []byte,
) (Media, error) {
	ret, e := NewMedia(media_t)
	if e != nil {
		return nil, e
	}

	e = ret.DeSerialize(pb_data)

	return ret, e
}

type Media interface {
	Type() pb.Media_Type

	DeSerialize(bs []byte) error
	Serialize() []byte

	FillFromJson(*ajson.Json) error
	ToJson() *ajson.Json

	FillMessage(m *pb.Message)

	// for send message
	MsgCategory() string // `text`/`media`

	// for encryption
	// for decryption
	SaltString() []byte
	MediaKey() []byte

	// for cdn
	MsgUrl() string
	DirectPath() string
	EncFileHash() []byte
	CdnHost(*def.Cdn) string
}

// placeholders for none-cdn-media like text/url...
type NoneCdnMedia struct {
}

func (*NoneCdnMedia) MsgUrl() string {
	return ``
}
func (*NoneCdnMedia) DirectPath() string {
	return ``
}
func (*NoneCdnMedia) EncFileHash() []byte {
	return nil
}
func (*NoneCdnMedia) CdnHost(cdn *def.Cdn) string {
	return ``
}

// placeholders for none-encrypted-media like text/url...
type NoneEncryptedMedia struct {
}

func (*NoneEncryptedMedia) SaltString() []byte {
	return nil
}
func (*NoneEncryptedMedia) MediaKey() []byte {
	return nil
}

// TodoMedia
type TodoMedia struct {
	NoneCdnMedia
	NoneEncryptedMedia
}

func (s *TodoMedia) Type() pb.Media_Type {
	return pb.Media_Unknown
}
func (s *TodoMedia) DeSerialize(bs []byte) error {
	return nil
}
func (s *TodoMedia) Serialize() []byte {
	return nil
}
func (s *TodoMedia) FillFromJson(*ajson.Json) error {
	return nil
}
func (s *TodoMedia) ToJson() *ajson.Json {
	return ajson.New()
}
func (s *TodoMedia) MsgCategory() string {
	return ``
}
func (s *TodoMedia) FillMessage(m *pb.Message) {
}

// text
type Text struct {
	NoneCdnMedia
	NoneEncryptedMedia
	Data []byte
}

func (t *Text) Serialize() []byte {
	if t.Data == nil {
		return []byte{}
	}
	return t.Data
}
func (t *Text) DeSerialize(bs []byte) error {
	t.Data = bs
	return nil
}
func (t *Text) FillFromJson(j *ajson.Json) error {
	t.Data = []byte(j.Get(`text`).String())
	return nil
}
func (t *Text) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`text`, string(t.Data))
	return j
}
func (t *Text) Type() pb.Media_Type {
	return pb.Media_Text
}
func (t *Text) MsgCategory() string {
	return `text`
}
func (t *Text) FillMessage(m *pb.Message) {
	m.Text = t.Data
}

// sticker
type Sticker struct {
	P *pb.Sticker
}

func (s *Sticker) Serialize() []byte {
	r, _ := proto.Marshal(s.P)
	return r
}
func (s *Sticker) DeSerialize(bs []byte) error {
	s.P = &pb.Sticker{}
	return proto.Unmarshal(bs, s.P)
}
func (s *Sticker) FillFromJson(j *ajson.Json) error {
	//s.P.MessageUrl = proto.String(j.Get(`messageUrl`).String())
	return nil
}
func (s *Sticker) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`messageUrl`, s.P.GetMessageUrl())
	j.Set(`fileHash`, algo.B64Enc(s.P.GetFileHash()))
	j.Set(`encFileHash`, algo.B64Enc(s.P.GetEncFileHash()))
	j.Set(`mediaKey`, algo.B64Enc(s.P.GetMediaKey()))
	j.Set(`mimeType`, s.P.GetMimeType())
	j.Set(`width`, s.P.GetWidth())
	j.Set(`height`, s.P.GetHeight())
	j.Set(`directPath`, s.P.GetDirectPath())
	j.Set(`fileLength`, s.P.GetFileLength())
	j.Set(`mediaKeyTimestamp`, s.P.GetMediaKeyTimestamp())
	return j
}
func (s *Sticker) MsgUrl() string {
	return s.P.GetMessageUrl()
}
func (s *Sticker) SaltString() []byte {
	return []byte(`WhatsApp Image Keys`)
}
func (s *Sticker) Type() pb.Media_Type {
	return pb.Media_Sticker
}
func (s *Sticker) MsgCategory() string {
	return `media`
}
func (s *Sticker) DirectPath() string {
	return s.P.GetDirectPath()
}
func (s *Sticker) EncFileHash() []byte {
	return s.P.GetEncFileHash()
}
func (s *Sticker) MediaKey() []byte {
	return s.P.GetMediaKey()
}
func (s *Sticker) CdnHost(cdn *def.Cdn) string {
	return cdn.Sticker
}
func (s *Sticker) FillMessage(m *pb.Message) {
	m.Sticker = s.P
}

// ptt
type Ptt struct {
	P *pb.Ptt
}

func (p *Ptt) Serialize() []byte {
	r, _ := proto.Marshal(p.P)
	return r
}
func (p *Ptt) DeSerialize(bs []byte) error {
	p.P = &pb.Ptt{}
	return proto.Unmarshal(bs, p.P)
}
func (p *Ptt) FillFromJson(j *ajson.Json) error {
	p.P.MessageUrl = proto.String(j.Get(`messageUrl`).String())
	p.P.MimeType = proto.String(j.Get(`mimeType`).String())
	if v, e := algo.B64Dec(j.Get(`fileHash`).String()); e == nil {
		p.P.FileHash = v
	}
	p.P.FileLength = proto.Uint32(uint32(j.Get(`fileLength`).Uint64()))
	p.P.MediaDuration = proto.Uint32(uint32(j.Get(`mediaDuration`).Uint64()))
	p.P.Origin = proto.Uint32(uint32(j.Get(`origin`).Uint64()))
	if v, e := algo.B64Dec(j.Get(`mediaKey`).String()); e == nil {
		p.P.MediaKey = v
	}
	if v, e := algo.B64Dec(j.Get(`encFileHash`).String()); e == nil {
		p.P.EncFileHash = v
	}
	p.P.DirectPath = proto.String(j.Get(`directPath`).String())
	p.P.MediaKeyTimestamp = proto.Uint32(uint32(j.Get(`mediaKeyTimestamp`).Uint64()))

	return nil
}
func (p *Ptt) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`messageUrl`, p.P.GetMessageUrl())
	j.Set(`mimeType`, p.P.GetMimeType())
	j.Set(`fileHash`, algo.B64Enc(p.P.GetFileHash()))
	j.Set(`fileLength`, p.P.GetFileLength())
	j.Set(`mediaDuration`, p.P.GetMediaDuration())
	j.Set(`origin`, p.P.GetOrigin())
	j.Set(`mediaKey`, algo.B64Enc(p.P.GetMediaKey()))
	j.Set(`encFileHash`, algo.B64Enc(p.P.GetEncFileHash()))
	j.Set(`directPath`, p.P.GetDirectPath())
	j.Set(`mediaKeyTimestamp`, p.P.GetMediaKeyTimestamp())
	return j
}
func (p *Ptt) MsgUrl() string {
	return p.P.GetMessageUrl()
}
func (p *Ptt) Type() pb.Media_Type {
	return pb.Media_Ptt
}
func (p *Ptt) MsgCategory() string {
	return `media`
}
func (p *Ptt) SaltString() []byte {
	return []byte(`WhatsApp Audio Keys`)
}
func (p *Ptt) DirectPath() string {
	return p.P.GetDirectPath()
}
func (p *Ptt) EncFileHash() []byte {
	return p.P.GetEncFileHash()
}
func (p *Ptt) MediaKey() []byte {
	return p.P.GetMediaKey()
}
func (p *Ptt) CdnHost(cdn *def.Cdn) string {
	return cdn.Ptt
}
func (p *Ptt) FillMessage(m *pb.Message) {
	m.Ptt = p.P
}

// image
type Image struct {
	P *pb.Image
}

func (i *Image) Serialize() []byte {
	r, _ := proto.Marshal(i.P)
	return r
}
func (i *Image) DeSerialize(bs []byte) error {
	i.P = &pb.Image{}
	return proto.Unmarshal(bs, i.P)
}
func (i *Image) FillFromJson(j *ajson.Json) error {
	i.P.MessageUrl = proto.String(j.Get(`messageUrl`).String())
	i.P.MimeType = proto.String(j.Get(`mimeType`).String())
	i.P.Text = []byte(j.Get(`text`).String())
	if v, e := algo.B64Dec(j.Get(`fileHash`).String()); e == nil {
		i.P.FileHash = v
	}
	i.P.FileLength = proto.Uint32(uint32(j.Get(`fileLength`).Uint64()))
	i.P.Height = proto.Uint32(uint32(j.Get(`height`).Uint64()))
	i.P.Width = proto.Uint32(uint32(j.Get(`width`).Uint64()))
	if v, e := algo.B64Dec(j.Get(`mediaKey`).String()); e == nil {
		i.P.MediaKey = v
	}
	if v, e := algo.B64Dec(j.Get(`encFileHash`).String()); e == nil {
		i.P.EncFileHash = v
	}
	i.P.DirectPath = proto.String(j.Get(`directPath`).String())
	i.P.MediaKeyTimestamp = proto.Uint32(uint32(j.Get(`mediaKeyTimestamp`).Uint64()))

	if v, e := algo.B64Dec(j.Get(`thumbnail`).String()); e == nil {
		i.P.Thumbnail = v
	}
	// sidecar
	// firstScanLength
	// partialMediaHash
	// partialMediaEHash

	return nil
}
func (i *Image) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`messageUrl`, i.P.GetMessageUrl())
	j.Set(`mimeType`, i.P.GetMimeType())
	j.Set(`text`, string(i.P.GetText()))
	j.Set(`fileHash`, algo.B64Enc(i.P.GetFileHash()))
	j.Set(`fileLength`, i.P.GetFileLength())
	j.Set(`height`, i.P.GetHeight())
	j.Set(`width`, i.P.GetWidth())
	j.Set(`mediaKey`, algo.B64Enc(i.P.GetMediaKey()))
	j.Set(`encFileHash`, algo.B64Enc(i.P.GetEncFileHash()))
	j.Set(`directPath`, i.P.GetDirectPath())
	j.Set(`mediaKeyTimestamp`, i.P.GetMediaKeyTimestamp())
	j.Set(`thumbnail`, algo.B64Enc(i.P.GetThumbnail()))
	return j
}
func (i *Image) MsgUrl() string {
	return i.P.GetMessageUrl()
}
func (i *Image) Type() pb.Media_Type {
	return pb.Media_Image
}
func (i *Image) MsgCategory() string {
	return `media`
}
func (i *Image) SaltString() []byte {
	return []byte(`WhatsApp Image Keys`)
}
func (i *Image) DirectPath() string {
	return i.P.GetDirectPath()
}
func (i *Image) EncFileHash() []byte {
	return i.P.GetEncFileHash()
}
func (i *Image) MediaKey() []byte {
	return i.P.GetMediaKey()
}
func (s *Image) CdnHost(cdn *def.Cdn) string {
	return cdn.Image
}
func (s *Image) FillMessage(m *pb.Message) {
	m.Image = s.P
}

// video
type Video struct {
	P *pb.Video
}

func (v *Video) Serialize() []byte {
	r, _ := proto.Marshal(v.P)
	return r
}
func (v *Video) DeSerialize(bs []byte) error {
	v.P = &pb.Video{}
	return proto.Unmarshal(bs, v.P)
}
func (v *Video) FillFromJson(j *ajson.Json) error {
	v.P.MessageUrl = proto.String(j.Get(`messageUrl`).String())
	v.P.MimeType = proto.String(j.Get(`mimeType`).String())
	v.P.Text = []byte(j.Get(`text`).String())
	if val, e := algo.B64Dec(j.Get(`fileHash`).String()); e == nil {
		v.P.FileHash = val
	}
	v.P.FileLength = proto.Uint32(uint32(j.Get(`fileLength`).Uint64()))
	v.P.MediaDuration = proto.Uint32(uint32(j.Get(`mediaDuration`).Uint64()))
	v.P.Height = proto.Uint32(uint32(j.Get(`height`).Uint64()))
	v.P.Width = proto.Uint32(uint32(j.Get(`width`).Uint64()))
	if val, e := algo.B64Dec(j.Get(`mediaKey`).String()); e == nil {
		v.P.MediaKey = val
	}
	if val, e := algo.B64Dec(j.Get(`encFileHash`).String()); e == nil {
		v.P.EncFileHash = val
	}
	v.P.DirectPath = proto.String(j.Get(`directPath`).String())
	v.P.MediaKeyTimestamp = proto.Uint32(uint32(j.Get(`mediaKeyTimestamp`).Uint64()))

	if val, e := algo.B64Dec(j.Get(`thumbnail`).String()); e == nil {
		v.P.Thumbnail = val
	}
	return nil
}
func (v *Video) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`messageUrl`, v.P.GetMessageUrl())
	j.Set(`mimeType`, v.P.GetMimeType())
	j.Set(`text`, string(v.P.GetText()))
	j.Set(`fileHash`, algo.B64Enc(v.P.GetFileHash()))
	j.Set(`mediaDuration`, v.P.GetMediaDuration())
	j.Set(`fileLength`, v.P.GetFileLength())
	j.Set(`height`, v.P.GetHeight())
	j.Set(`width`, v.P.GetWidth())
	j.Set(`mediaKey`, algo.B64Enc(v.P.GetMediaKey()))
	j.Set(`encFileHash`, algo.B64Enc(v.P.GetEncFileHash()))
	j.Set(`directPath`, v.P.GetDirectPath())
	j.Set(`mediaKeyTimestamp`, v.P.GetMediaKeyTimestamp())
	j.Set(`thumbnail`, algo.B64Enc(v.P.GetThumbnail()))
	return j
}
func (v *Video) MsgUrl() string {
	return v.P.GetMessageUrl()
}
func (v *Video) Type() pb.Media_Type {
	return pb.Media_Video
}
func (v *Video) MsgCategory() string {
	return `media`
}
func (v *Video) SaltString() []byte {
	return []byte(`WhatsApp Video Keys`)
}
func (v *Video) DirectPath() string {
	return v.P.GetDirectPath()
}
func (v *Video) EncFileHash() []byte {
	return v.P.GetEncFileHash()
}
func (v *Video) MediaKey() []byte {
	return v.P.GetMediaKey()
}
func (s *Video) CdnHost(cdn *def.Cdn) string {
	return cdn.Video
}
func (s *Video) FillMessage(m *pb.Message) {
	m.Video = s.P
}

// Url
type Url struct {
	NoneCdnMedia
	NoneEncryptedMedia

	P *pb.Url
}

func (u *Url) Serialize() []byte {
	r, _ := proto.Marshal(u.P)
	return r
}
func (u *Url) DeSerialize(bs []byte) error {
	u.P = &pb.Url{}
	return proto.Unmarshal(bs, u.P)
}
func (u *Url) FillFromJson(j *ajson.Json) error {
	u.P.Data = proto.String(j.Get(`data`).String())
	if v, e := j.Get(`url`).TryString(); e == nil {
		u.P.Url = proto.String(v)
	}
	if v, e := j.Get(`mediaName`).TryString(); e == nil {
		u.P.MediaName = proto.String(v)
	}
	if v, e := j.Get(`mediaCaption`).TryString(); e == nil {
		u.P.MediaCaption = proto.String(v)
	}
	if v, e := j.Get(`textColor`).TryUint64(); e == nil {
		u.P.TextColor = proto.Uint32(uint32(v))
	}
	if v, e := j.Get(`backgroundColor`).TryUint64(); e == nil {
		u.P.BackgroundColor = proto.Uint32(uint32(v))
	}
	if v, e := j.Get(`fontStyle`).TryInt(); e == nil {
		u.P.FontStyle = proto.Int32(int32(v))
	}

	u.P.Int_10 = proto.Int32(0)
	if v, e := j.Get(`thumbImage`).TryString(); e == nil {
		var e error
		u.P.ThumbImage, e = algo.B64Dec(v)
		if e != nil {
			return e
		}
	}
	if v, e := j.Get(`directPath`).TryString(); e == nil {
		u.P.DirectPath = proto.String(v)
	}
	if v, e := j.Get(`imgHeight`).TryInt(); e == nil {
		u.P.ImgHeight = proto.Int32(int32(v))
	}
	if v, e := j.Get(`imgWidth`).TryInt(); e == nil {
		u.P.ImgWidth = proto.Int32(int32(v))
	}

	return nil
}
func (u *Url) ToJson() *ajson.Json {
	j := ajson.New()
	if u.P.Data != nil {
		j.Set(`data`, u.P.GetData())
	}
	if u.P.Url != nil {
		j.Set(`url`, u.P.GetUrl())
	}
	if u.P.MediaName != nil {
		j.Set(`mediaName`, u.P.GetMediaName())
	}
	if u.P.MediaCaption != nil {
		j.Set(`mediaCaption`, u.P.GetMediaCaption())
	}
	if u.P.TextColor != nil {
		j.Set(`textColor`, u.P.GetTextColor())
	}
	if u.P.BackgroundColor != nil {
		j.Set(`backgroundColor`, u.P.GetBackgroundColor())
	}
	if u.P.FontStyle != nil {
		j.Set(`fontStyle`, u.P.GetFontStyle())
	}

	if u.P.ThumbImage != nil {
		j.Set(`thumbImage`, algo.B64Enc(u.P.GetThumbImage()))
	}
	if u.P.DirectPath != nil {
		j.Set(`directPath`, u.P.GetDirectPath())
	}
	if u.P.ImgHeight != nil {
		j.Set(`imgHeight`, u.P.GetImgHeight())
	}
	if u.P.ImgWidth != nil {
		j.Set(`imgWidth`, u.P.GetImgWidth())
	}
	return j
}
func (u *Url) Type() pb.Media_Type {
	return pb.Media_Url
}
func (u *Url) MsgCategory() string {
	return `media`
}
func (u *Url) FillMessage(m *pb.Message) {
	m.Url = u.P
}

// Document
type Document struct {
	P *pb.Document
}

func (x *Document) Serialize() []byte {
	r, _ := proto.Marshal(x.P)
	return r
}
func (x *Document) DeSerialize(bs []byte) error {
	x.P = &pb.Document{}
	return proto.Unmarshal(bs, x.P)
}
func (x *Document) FillFromJson(j *ajson.Json) error {
	x.P.MediaUrl = proto.String(j.Get(`mediaUrl`).String())
	x.P.MimeType = proto.String(j.Get(`mimeType`).String())
	if v, e := algo.B64Dec(j.Get(`mediaHash`).String()); e == nil {
		x.P.MediaHash = v
	}
	x.P.MediaSize = proto.Uint32(uint32(j.Get(`mediaSize`).Uint64()))
	x.P.Int_6 = proto.Int32(0)
	if v, e := algo.B64Dec(j.Get(`mediaKey`).String()); e == nil {
		x.P.MediaKey = v
	}
	if v, e := algo.B64Dec(j.Get(`mediaEncHash`).String()); e == nil {
		x.P.MediaEncHash = v
	}
	x.P.DirectPath = proto.String(j.Get(`directPath`).String())
	x.P.MediaKeyTimestamp = proto.Uint32(uint32(j.Get(`mediaKeyTimestamp`).Uint64()))

	return nil
}
func (x *Document) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`mediaUrl`, x.P.GetMediaUrl())
	j.Set(`mimeType`, x.P.GetMimeType())
	j.Set(`mediaName`, x.P.GetMediaName())
	j.Set(`mediaHash`, algo.B64Enc(x.P.GetMediaHash()))
	j.Set(`mediaSize`, x.P.GetMediaSize())
	j.Set(`int_6`, x.P.GetInt_6())
	j.Set(`mediaKey`, algo.B64Enc(x.P.GetMediaKey()))
	j.Set(`mediaCaption`, x.P.GetMediaCaption())
	j.Set(`mediaEncHash`, algo.B64Enc(x.P.GetMediaEncHash()))
	j.Set(`directPath`, x.P.GetDirectPath())
	j.Set(`mediaKeyTimestamp`, x.P.GetMediaKeyTimestamp())

	return j
}
func (x *Document) MsgUrl() string {
	return x.P.GetMediaUrl()
}
func (x *Document) Type() pb.Media_Type {
	return pb.Media_Document
}
func (x *Document) MsgCategory() string {
	return `media`
}
func (x *Document) SaltString() []byte {
	return []byte(`WhatsApp Document Keys`)
}
func (x *Document) DirectPath() string {
	return x.P.GetDirectPath()
}
func (x *Document) EncFileHash() []byte {
	return x.P.GetMediaEncHash()
}
func (x *Document) MediaKey() []byte {
	return x.P.GetMediaKey()
}
func (x *Document) CdnHost(cdn *def.Cdn) string {
	return cdn.Document
}
func (x *Document) FillMessage(m *pb.Message) {
	m.Document = x.P
}

// Contact
type Contact struct {
	NoneCdnMedia
	NoneEncryptedMedia
	P *pb.Contact
}

func (c *Contact) Serialize() []byte {
	r, _ := proto.Marshal(c.P)
	return r
}
func (c *Contact) DeSerialize(bs []byte) error {
	c.P = &pb.Contact{}
	return proto.Unmarshal(bs, c.P)
}
func (c *Contact) FillFromJson(j *ajson.Json) error {
	c.P.Name = proto.String(j.Get(`name`).String())
	c.P.Vcard = proto.String(j.Get(`vcard`).String())
	return nil
}
func (c *Contact) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`name`, c.P.GetName())
	j.Set(`vcard`, c.P.GetVcard())
	return j
}
func (c *Contact) Type() pb.Media_Type {
	return pb.Media_Contact
}
func (c *Contact) MsgCategory() string {
	return `media`
}
func (c *Contact) FillMessage(m *pb.Message) {
	m.Contact = c.P
}

// ContactArray
type ContactArray struct {
	NoneCdnMedia
	NoneEncryptedMedia
	P *pb.ContactArray
}

func (ca *ContactArray) Serialize() []byte {
	r, _ := proto.Marshal(ca.P)
	return r
}
func (ca *ContactArray) DeSerialize(bs []byte) error {
	ca.P = &pb.ContactArray{}
	return proto.Unmarshal(bs, ca.P)
}
func (ca *ContactArray) FillFromJson(j *ajson.Json) error {
	ca.P.Title = proto.String(j.Get(`title`).String())
	for _, c := range j.Get(`list`).JsonArray() {
		ca.P.List = append(ca.P.List, &pb.Contact{
			Name:  proto.String(c.Get(`name`).String()),
			Vcard: proto.String(c.Get(`vcard`).String()),
		})
	}
	return nil
}
func (ca *ContactArray) ToJson() *ajson.Json {
	j := ajson.New()
	j.Set(`title`, ca.P.GetTitle())
	for _, c := range ca.P.GetList() {
		x := ajson.New()
		x.Set(`name`, c.GetName())
		x.Set(`vcard`, c.GetVcard())

		j.Add(`list`, x)
	}
	return j
}
func (ca *ContactArray) Type() pb.Media_Type {
	return pb.Media_Contact_Array
}
func (ca *ContactArray) MsgCategory() string {
	return `media`
}
func (ca *ContactArray) FillMessage(m *pb.Message) {
	m.ContactArray = ca.P
}

func encryptMedia(
	media_key, salt, raw_file []byte,
) ([]byte, error) {
	if len(media_key) != 0x20 {
		return nil, errors.New(`invalid media key`)
	}
	x, e := algo.HkdfSha256(
		media_key,
		make([]byte, 0x20), // 00000000000000000000000000000000
		salt,
		0x50,
	)
	if e != nil {
		return nil, e
	}
	iv := x[0:0x10]
	key := x[0x10:0x30]
	mac_key := x[0x30:]

	enc, e := algo.AesCbcPkcsEncrypt(raw_file, key, iv)
	if e != nil {
		return nil, e
	}
	mac := hmac.New(sha256.New, mac_key)
	mac.Write(iv)
	mac.Write(enc)
	mc := mac.Sum(nil)

	all := append(enc, mc[0:10]...)
	return all, nil
}
func decryptMedia(m Media, bs []byte) ([]byte, error) {
	if len(bs) < 10 {
		return nil, errors.New(`data too short`)
	}

	x, e := algo.HkdfSha256(
		m.MediaKey(),
		make([]byte, 0x20), // 00000000000000000000000000000000
		m.SaltString(), 0x30,
	)
	if e != nil {
		return nil, e
	}
	iv := x[:0x10]
	key := x[0x10:]

	cipher := bs[:len(bs)-10] // except last 10 bytes Mac

	//return algo.AesCbcDecrypt(cipher, key, iv, &algo.None{})
	return algo.AesCbcPkcsDecrypt(cipher, key, iv)
}
