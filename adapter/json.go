package adapter

import (
	"strconv"

	"github.com/darabuchi/log"
	"github.com/valyala/fastjson"
)

func getInt(j *fastjson.Value, field string) int {
	val := j.Get(field)
	if val == nil {
		return 0
	}

	switch val.Type() {
	case fastjson.TypeString:
		i, err := strconv.Atoi(string(j.GetStringBytes(field)))
		if err != nil {
			return 0
		}

		return i
	case fastjson.TypeNumber:
		i, err := val.Int()
		if err != nil {
			return 0
		}
		return i
	default:
		log.Warnf("%s type is %s", field, val.Type())
		return 0
	}
}
