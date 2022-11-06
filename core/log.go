package core

func NewLogHook(a *Acc) func(...any) error {
	return func(args ...any) error {
		lvl := args[0].(int)
		format := args[1].(string)

		// convert []any.Any -> []interface{}
		params := []interface{}{}
		for _, v := range args[2:] {
			var x interface{}
			x = v
			params = append(params, x)
		}
		a.Log.DoLog(lvl, format, params...)

		return nil
	}
}
