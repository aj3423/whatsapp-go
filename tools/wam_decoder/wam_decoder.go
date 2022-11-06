package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"ahex"
	"wa/wam"

	"github.com/fatih/color"
)

func main() {
	var bs []byte
	if len(os.Args) == 2 {
		bs = ahex.Dec(os.Args[1])
	} else {
		color.HiRed("usage: wam 010203....")
		os.Exit(1)
	}

	recs, e := wam.Parse(bs)
	if e != nil {
		color.HiRed(e.Error())
		os.Exit(1)
	}

	var cur_fid int32 = 0

	for _, rec := range recs {
		idStr := color.HiBlueString("%d", rec.Id)

		if rec.ClassType == 1 { // class begin
			fmt.Println(``)
			if ms, ok := wam.MapStats[rec.Id]; ok {
				idStr = color.HiGreenString("%d\t(%s)", rec.Id, ms.Desc)
			}

			cur_fid = rec.Id
		} else if fa, ok := wam.MapStats[cur_fid]; ok {
			if statName, ok := fa.Stats[rec.Id]; ok {
				idStr += color.HiBlueString("\t(%s)", statName)
			}
		}

		valStr := color.HiYellowString("%v", rec.Value)

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

		all := fmt.Sprintf("%d %s %s %s  ",
			rec.ClassType,
			idStr,
			valStr,
			valType,
		)
		if rec.IsClassEnd { // chunk End
			all += color.HiMagentaString("<- chunk end\n")
			cur_fid = 0
		}

		fmt.Println(all)
	}
}
