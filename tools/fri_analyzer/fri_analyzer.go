package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"acolor"
	"aconfig"
	"afs"
	"ahex"
	"algo"
	"aregex"
	"buffer"
	"gui"
	"wa/wam"
	"wa/xmpp"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
)

var CFG_File = "config.toml"

type Config struct {
	LastFile string
}

var cfg = &Config{}

func init() {
	aconfig.Load(CFG_File, cfg)
}

const time_fmt string = "01/02 15:04:05.000"

var master *giu.MasterWindow

const (
	Mode_Timeline = 0
	Mode_Relation = 1
)

const (
	Type_Enc = iota
	Type_Dec
	Type_WamRate
	Type_WamExecute
)

const (
	Filter_Mode_Hide      = true
	Filter_Mode_Highlight = false
)

var (
	watch_file bool
	watcher    *afs.Watcher

	mode              int = Mode_Relation
	drop_file         string
	filter            string
	filter_mode       bool = Filter_Mode_Highlight
	with_ping         bool = false
	with_routing_info bool = false
	with_unhit_wam    bool = false
)

type SampleRate struct {
	a1   string
	a2   string
	rate string
}
type WamRate struct {
	Code       int32  // 854
	SampleRate        // <1,20,20>
	HitResult  string // null / 20
}
type WamExecute struct {
	Channel int32 // 0, 1, 2
	Code    int32 // 854
	Weight  int32
}

type RowData struct {
	Type       int       // enc/dec
	Time       time.Time // timestamp from "8/9 21:12:5.123"
	*xmpp.Node           // 00f8..

	Highlight bool // filtered
	Resp      *RowData
	*WamRate
	*WamExecute
}

var sel_index = -1

func render_id_column(rd *RowData) string {
	if rd.Node != nil && rd.Node.Tag == `iq` {
		id, ok := rd.Node.GetAttr(`id`)
		if !ok {
			id = ``
		}
		return id
	} else {
		return ``
	}
}

func render_req_column(rd *RowData, col_type int) string {
	switch col_type {
	case Type_Enc:
		if rd.Type == Type_Enc {
			l := rd.Node.Tag
			type_, ok_1 := rd.Node.GetAttr(`type`)
			xmlns, ok_2 := rd.Node.GetAttr(`xmlns`)
			if ok_1 {
				l += "." + type_
			}
			if ok_2 {
				l += "." + xmlns
			}
			return l
		}

	case Type_Dec:
		switch rd.Type {
		case Type_Enc:
			if rd.Resp != nil {
				return rd.Resp.Node.Tag
			}
		case Type_Dec:
			return rd.Node.Tag
		}
	}
	return ``
}
func render_wam_column(rd *RowData) string {
	if ev := rd.WamRate; ev != nil {
		// if not found in attr map, just show number
		code_name := strconv.Itoa(int(ev.Code))
		// if found in attr map, show name
		if cls_attr, ok := wam.MapStats[ev.Code]; ok {
			code_name = fmt.Sprintf("%d %s", ev.Code, cls_attr.Desc)
		}
		return fmt.Sprintf("%s: %s", code_name, ev.HitResult)
	}
	if exec := rd.WamExecute; exec != nil {
		// if not found in attr map, just show number
		code_name := strconv.Itoa(int(exec.Code))
		// if found in attr map, show name
		if cls_attr, ok := wam.MapStats[exec.Code]; ok {
			code_name = fmt.Sprintf("%d %s", exec.Code, cls_attr.Desc)
		}
		return fmt.Sprintf("[%d]: %s, %d", exec.Channel, code_name, exec.Weight)
	}
	return ``
}
func render_wam_rate_tip(rd *RowData) string {
	if rd.Type == Type_WamRate {
		return fmt.Sprintf("<%s, %s, %s>",
			rd.WamRate.a1, rd.WamRate.a2, rd.WamRate.rate)
	}
	return ``
}

var mu_rows sync.RWMutex
var all_rows []*RowData // all rows in file
var rows []*RowData     // only filtered, for display

func bytes_2_node(b []byte) (*xmpp.Node, error) {
	if len(b) < 2 {
		return nil, errors.New(`data too short: ` + ahex.Enc(b))
	}
	if b[0] == 0 && b[1] == 0xf8 {
		b = b[1:]
	} else if b[0] == 2 && b[1] == 0x78 && b[2] == 0x9c {
		dec, e := algo.UnZlib(b[1:])
		if e != nil {
			return nil, e
		}
		return bytes_2_node(dec)
	}
	r := xmpp.NewReader(b)

	return r.ReadNode()
}

func sort_rows() {
	mu_rows.Lock()
	defer mu_rows.Unlock()

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Time.Before(rows[j].Time)
	})
}

// useless, just string holder for InputTextMultiline
var txt_req string
var wam_req string
var txt_resp string

func render_req() (giu.Widget, bool) {
	if sel_index >= 0 && sel_index < len(rows) {
		rd := rows[sel_index]

		switch rd.Type {
		case Type_Enc:
			txt_req = rd.Node.ToString()

			// wam
			if tag, _ := rd.Node.GetAttr(`xmlns`); tag == `w:stats` {
				// parse wam
				{
					if ch, ok := rd.Node.FindChildByTag(`add`); ok {
						recs, e := wam.Parse(ch.Data)
						if e != nil {
							fmt.Println(e)
						} else {
							wam_req = dump_wam(recs)
						}
					}
				}

				return giu.InputTextMultiline(&txt_req).Size(giu.Auto, giu.Auto),
					true
			}
		default:
			txt_req = ``
		}
	} else {
		txt_req = ``
	}
	return giu.InputTextMultiline(&txt_req).Size(giu.Auto, giu.Auto), false
}
func render_resp() giu.Widget {
	if sel_index >= 0 && sel_index < len(rows) {
		rd := rows[sel_index]

		switch rd.Type {
		case Type_Enc:
			if rd.Resp == nil {
				txt_resp = ``
			} else {
				txt_resp = rd.Resp.Node.ToString()
			}
		case Type_Dec:
			txt_resp = rd.Node.ToString()
		}
	} else {
		txt_resp = ``
	}
	return giu.InputTextMultiline(&txt_resp).Size(giu.Auto, giu.Auto)
}

func on_row_click(index int) {
	rd := rows[index]
	switch rd.Type {
	case Type_Enc:
		txt_req = rd.Node.ToString()
		if rd.Resp != nil {
			txt_resp = rd.Resp.Node.ToString()
		} else {
			txt_resp = ``
		}
	case Type_Dec:
		txt_req = ``
		txt_resp = rd.Node.ToString()
	}
}
func format_duration(dur time.Duration) string {
	return fmt.Sprintf("      +  %s", dur)
}
func render_rows() []*giu.TableRowWidget {
	mu_rows.RLock()
	defer mu_rows.RUnlock()

	ret := []*giu.TableRowWidget{}

	prev_t := time.Time{}

	for i, row := range rows {
		index := i

		id := render_id_column(rows[index])

		// time column
		tm_full := row.Time.Format(time_fmt)
		var tm_offset string
		if !prev_t.IsZero() && prev_t.Add(5*time.Minute).After(row.Time) { // if time in 5 minutes
			tm_offset = format_duration(row.Time.Sub(prev_t))
		}
		prev_t = row.Time

		req := render_req_column(row, Type_Enc)
		resp := render_req_column(row, Type_Dec)
		wam_rate := render_wam_column(row)
		wam_rate_tip := render_wam_rate_tip(row)

		rwgt := giu.TableRow(
			giu.Custom(func() {
				giu.Selectable(id).Selected(sel_index == index).
					Flags(giu.SelectableFlagsSpanAllColumns | giu.SelectableFlagsAllowDoubleClick).Build()

				// row click
				if giu.IsItemActive() && giu.IsMouseClicked(giu.MouseButtonLeft) {
					sel_index = index
				}
			}),
			giu.Custom(func() {
				if tm_offset != `` {
					giu.Label(tm_offset).Build()
					giu.Tooltip(tm_full).Build()
				} else {
					giu.Label(tm_full).Build()
				}
			}),
			giu.Custom(func() {
				// render w:stats with yellow color
				if req == `iq.set.w:stats` {
					imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.3, Y: 0.3, Z: 1, W: 1})
					defer imgui.PopStyleColor()
				}
				giu.Label(req).Build()
			}),
			giu.Custom(func() {
				if resp == `success` {
					imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.3, Y: 0.8, Z: 0.8, W: 1})
					defer imgui.PopStyleColor()
				}

				giu.Label(resp).Build()
			}),
			giu.Custom(func() {
				//if strings.Contains(wam_rate, `WamAppLaunch`) {
				//imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.8, Y: 0.8, Z: 0.3, W: 1})
				//defer imgui.PopStyleColor()
				//}
				giu.Label(wam_rate).Build()
				if len(wam_rate_tip) > 0 {
					giu.Tooltip(wam_rate_tip).Build()
				}
			}),
		)
		if row.Highlight {
			rwgt.BgColor(&acolor.Peru)
		}
		ret = append(ret, rwgt)
	}

	if len(ret) == 0 {
		ret = append(ret, giu.TableRow())
	}
	return ret
}

func parse_req_resp_rows_from_regex(html, pattern string, enc_type int) []*RowData {
	var ret []*RowData

	r := aregex.SearchAll(html, pattern)

	for _, enc := range r {
		t, e := time.Parse(time_fmt, string(enc[1]))
		if e != nil {
			panic(e)
		}
		if n, e := bytes_2_node(ahex.Dec(string(enc[2]))); e == nil {
			rd := &RowData{
				Type: enc_type,
				Time: t,
				Node: n,
			}
			ret = append(ret, rd)
		}
	}
	return ret
}
func parse_wam_get_weight_from_regex(html, pattern string) []*RowData {
	var ret []*RowData

	r := aregex.SearchAll(html, pattern)

	for _, m := range r {
		t, e := time.Parse(time_fmt, string(m[1]))
		if e != nil {
			panic(e)
		}

		code, e := strconv.Atoi(string(m[2]))
		if e != nil {
			panic(e)
		}
		ret = append(ret, &RowData{
			Type: Type_WamRate,
			Time: t,
			WamRate: &WamRate{
				Code: int32(code),
				SampleRate: SampleRate{
					a1:   string(m[3]),
					a2:   string(m[4]),
					rate: string(m[5]),
				},
				HitResult: string(m[6]),
			},
		})
	}
	return ret
}
func parse_wam_execute_from_regex(html, pattern string) []*RowData {
	var ret []*RowData

	r := aregex.SearchAll(html, pattern)

	for _, m := range r {
		t, e := time.Parse(time_fmt, string(m[1]))
		if e != nil {
			panic(e)
		}

		channel, e := strconv.Atoi(string(m[2]))
		if e != nil {
			panic(e)
		}
		code, e := strconv.Atoi(string(m[3]))
		if e != nil {
			panic(e)
		}
		weight, e := strconv.Atoi(string(m[4]))
		if e != nil {
			panic(e)
		}
		ret = append(ret, &RowData{
			Type: Type_WamExecute,
			Time: t,
			WamExecute: &WamExecute{
				Channel: int32(channel),
				Code:    int32(code),
				Weight:  int32(weight),
			},
		})
	}
	return ret
}

func find_req_from_resp(resp *RowData) *RowData {
	iq_id, ok := resp.GetAttr(`id`)
	if !ok {
		return nil
	}

	for _, req := range rows {
		if req.Resp != nil {
			continue
		}
		if id, ok := req.GetAttr(`id`); ok {
			if id == iq_id && req.Time.Before(resp.Time) { // req.Time < resp.Time
				return req
			}
		}
	}
	return nil
}

func process_file() {

	master.SetTitle(cfg.LastFile)

	//sel_index = -1

	f := afs.ReadStr(cfg.LastFile)

	mu_rows.Lock()

	// parse Encrypt
	reqs := parse_req_resp_rows_from_regex(f, `<(.+?)>\r?\n---- AesGcmWrap.Encrypt: ----\r?\ndata:(.+)\r?\n`, Type_Enc)

	rows = reqs

	// parse Decrypt
	resps := parse_req_resp_rows_from_regex(f, `<(.+?)>\r?\n---- AesGcmWrap.Decrypt: ----\r?\nresult:(.+)\r?\n`, Type_Dec)
	for _, resp := range resps {
		switch mode {
		case Mode_Timeline:
			rows = append(rows, resp)

		case Mode_Relation:
			if req := find_req_from_resp(resp); req != nil {
				req.Resp = resp
			} else {
				rows = append(rows, resp)
			}
		}
	}

	// wam hit rate
	rates := parse_wam_get_weight_from_regex(f, `<(.+?)> <get_weight>: (.+?), <(.+?), (.+?), (.+?)>: (.+?)\r?\n`)
	rows = append(rows, rates...)

	// wam hit rate
	execs := parse_wam_execute_from_regex(f, `<(.+?)> <execute>: (.+?), (.+?), (.+?)\r?\n`)
	rows = append(rows, execs...)

	mu_rows.Unlock()

	sort_rows()

	all_rows = rows // backup for filter change

	filter_rows()
}
func refresh() {
	process_file()
}

func dump_wam(recs []*wam.Record) string {

	out := &strings.Builder{}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	var cur_fid int32 = 0

	for _, rec := range recs {
		idStr := fmt.Sprintf("%d", rec.Id)

		if rec.ClassType == 1 { // class begin
			fmt.Fprintln(w)
			if ms, ok := wam.MapStats[rec.Id]; ok {
				idStr = fmt.Sprintf("%d\t<%s>", rec.Id, ms.Desc)
			}

			cur_fid = rec.Id
		} else if fa, ok := wam.MapStats[cur_fid]; ok {
			if statName, ok := fa.Stats[rec.Id]; ok {
				idStr += fmt.Sprintf("\t(%s)", statName)
			}
		}

		valStr := fmt.Sprintf("%v", rec.Value)

		valType := `nil`
		if rec.Value != nil {
			valType = reflect.TypeOf(rec.Value).String()
			if cur_fid == 0 && rec.Id == 47 {
				if val, ok := rec.Value.(int32); ok {
					tm := time.Unix(int64(val), 0)
					valStr += fmt.Sprintf(" (%s)", tm.Format("2006-01-02 15:04:05"))
				}
			}
		}

		all := fmt.Sprintf("%d\t%s\t%s\t%s  ",
			rec.ClassType,
			idStr,
			valType,
			valStr,
		)
		if rec.IsClassEnd { // chunk End
			all += fmt.Sprintf("<- chunk end\n")
			cur_fid = 0
		}

		fmt.Fprintln(w, all)
	}
	return out.String()
}

func is_ping(n *xmpp.Node, t int) bool {
	if n == nil {
		return false
	}

	if t == Type_Enc {
		if n.Tag == `iq` {
			attrs := n.MapAttrs()

			// 1. my ping ->
			{
				to, _ := attrs[`to`]
				type_, _ := attrs[`type`]
				xmlns, _ := attrs[`xmlns`]
				if to == `s.whatsapp.net` && type_ == `get` && xmlns == `w:p` {
					if _, ok := n.FindChildByTag(`ping`); ok {
						return true
					}
				}
			}
			// 2. server ping ack ->
			{
				type_, _ := attrs[`type`]
				to, _ := attrs[`to`]
				if n.Children == nil && len(attrs) == 2 && type_ == `result` && to == `s.whatsapp.net` {
					return true
				}
			}
		}
	} else if t == Type_Dec {
		if n.Tag == `iq` {
			attrs := n.MapAttrs()

			// 1. my ping ack <-
			{
				_, has_t := attrs[`t`]
				type_, _ := attrs[`type`]
				from, _ := attrs[`from`]
				if n.Children == nil && len(attrs) == 4 && has_t && type_ == `result` && from == `s.whatsapp.net` {
					return true
				}
			}
			// 2. server ping <-
			{
				_, has_t := attrs[`t`]
				type_, _ := attrs[`type`]
				xmlns, _ := attrs[`xmlns`]
				if n.Children == nil && len(attrs) == 4 && has_t && type_ == `get` && xmlns == `urn:xmpp:ping` {
					return true
				}
			}
		}
	}
	return false
}
func is_routing_info(n *xmpp.Node, t int) bool {
	if t == Type_Dec {
		if n.Tag == `ib` && len(n.Attrs) == 1 && len(n.Children) == 1 {
			if from, _ := n.GetAttr(`from`); from == `s.whatsapp.net` {
				if ch := n.Children[0]; ch.Tag == `edge_routing` {
					return true
				}
			}
		}
	}
	return false
}

// filter from `all_rows`, save to `rows`
func filter_rows() {
	mu_rows.Lock()
	defer mu_rows.Unlock()

	rows = []*RowData{}
	for _, row := range all_rows {

		id := render_id_column(row)
		time := row.Time.Format(time_fmt)

		var req string
		if row.Node != nil {
			req = row.Node.ToString()
		}
		var resp string
		if row.Resp != nil {
			resp = row.Resp.Node.ToString()
		}

		// filter ping(s)
		if !with_ping {
			if is_ping(row.Node, row.Type) {
				continue
			}
		}
		if !with_routing_info {
			if is_routing_info(row.Node, row.Type) {
				continue
			}
		}

		// filter Wam
		var wam_dump string
		if row.Type == Type_Enc && row.Node != nil {
			if tag, _ := row.Node.GetAttr(`xmlns`); tag == `w:stats` {
				if ch, ok := row.Node.FindChildByTag(`add`); ok {
					recs, e := wam.Parse(ch.Data)
					if e == nil {
						wam_dump = dump_wam(recs)
					}
				}
			}
		}

		// filter Wam Rate
		var wam_rate string
		if row.Type == Type_WamRate {
			wam_rate = render_wam_column(row)
		}

		// filter unhit Wam Rate
		if !with_unhit_wam && row.Type == Type_WamRate {
			if row.HitResult == `0` {
				continue
			}
		}

		row.Highlight = false

		Lower := strings.ToLower
		hit_filter := false
		if len(filter) > 0 {
			lower := Lower(filter)
			hit_filter = strings.Contains(Lower(id), lower) ||
				strings.Contains(Lower(time), lower) ||
				strings.Contains(Lower(req), lower) ||
				strings.Contains(Lower(resp), lower) ||
				strings.Contains(Lower(wam_dump), lower) ||
				strings.Contains(Lower(wam_rate), lower)
		}

		switch filter_mode {
		case Filter_Mode_Hide:
			if hit_filter || len(filter) == 0 {
				rows = append(rows, row)
			}
		case Filter_Mode_Highlight:
			if hit_filter {
				row.Highlight = true
			}
			rows = append(rows, row)
		}
	}
}

var input_filter *gui.FocusableWidget
var timed_buffer *buffer.Timed
var file_chaning bool = false

func loop() {
	r, is_wam := render_req()
	giu.SingleWindow().Layout(
		giu.Row(
			giu.Button(`open last`).OnClick(func() {
				process_file()
			}),
			giu.Custom(func() {
				if file_chaning {
					imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.3, Y: 0.3, Z: 1, W: 1})
					defer imgui.PopStyleColor()
				}

				giu.Checkbox("Watch File", &watch_file).OnChange(func() {
					if watch_file {

						watcher, _ = afs.NewFileWatcher(cfg.LastFile, &afs.WatchOption{
							OnModify: func() {
								file_chaning = true
								timed_buffer.Buffer(func() {
									process_file()
									file_chaning = false
								})()
							},
						})
					} else {
						if watcher != nil {
							watcher.Stop()
						}
					}
				}).Build()
			}),
		),
		giu.Row(
			giu.RadioButton("Timeline", mode == Mode_Timeline).OnChange(func() {
				mode = Mode_Timeline
				process_file()
			}),
			giu.RadioButton("Relation", mode == Mode_Relation).OnChange(func() {
				mode = Mode_Relation
				process_file()
			}),
			giu.Checkbox(`Ping`, &with_ping).OnChange(filter_rows),
			giu.Checkbox(`RoutingInfo`, &with_routing_info).OnChange(filter_rows),
			giu.Checkbox(`UnHit Wam`, &with_unhit_wam).OnChange(filter_rows),
			input_filter,
			giu.Checkbox(`Filter Hide Mode`, &filter_mode).OnChange(filter_rows),
		),
		giu.SplitLayout(giu.DirectionHorizontal, &f740,
			giu.Table().Size(-1, -1).FastMode(true).
				Columns(
					giu.TableColumn(`ID`).Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(20),
					giu.TableColumn(`Time`).Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(150),
					giu.TableColumn(`Req`).Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(180),
					giu.TableColumn(`Resp`).Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(60),
					giu.TableColumn(`WamHit`).Flags(giu.TableColumnFlagsWidthFixed).InnerWidthOrWeight(300),
				).
				Rows(render_rows()...),
			giu.Custom(func() {
				req_resp := giu.SplitLayout(giu.DirectionVertical, &f400,
					r,
					render_resp(),
				)

				if is_wam {
					giu.SplitLayout(giu.DirectionHorizontal, &f1000,
						giu.InputTextMultiline(&wam_req).Size(-1, -1),
						req_resp,
					).Build()
				} else {
					req_resp.Build()
				}
			}),
		),
	)
}

func on_drop(names []string) {
	if len(names) != 1 {
		giu.OpenPopup(`Confirm`)
	}
	cfg.LastFile = names[0]
	aconfig.Save(CFG_File, cfg)
	process_file()
	giu.Update()
}

var f740 float32 = 340
var f400 float32 = 400
var f1000 float32 = 1000

func main() {
	// giu.SetDefaultFont("Fira Code", 16)

	timed_buffer = buffer.NewTimed(time.Second)

	master = giu.NewMasterWindow("", 1300, 700, 0).
		RegisterKeyboardShortcuts(
			giu.WindowShortcut{
				Key:      giu.KeyW,
				Modifier: giu.ModControl,
				Callback: func() { os.Exit(0) },
			},
			giu.WindowShortcut{
				Key:      giu.KeyR,
				Modifier: giu.ModControl,
				Callback: refresh,
			},
			giu.WindowShortcut{
				Key:      giu.KeyF,
				Modifier: giu.ModControl,
				Callback: func() { input_filter.Focus() },
			},
			giu.WindowShortcut{
				Key:      giu.KeyO,
				Modifier: giu.ModControl,
				Callback: process_file,
			},
		)

	input_filter = gui.Focusable(giu.InputText(&filter).Size(200).OnChange(filter_rows))

	master.SetDropCallback(on_drop)

	master.Run(loop)
}
