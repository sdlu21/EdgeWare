package sample

import "github.com/project-flogo/core/data/coerce"

type Input struct {
	DisplayJson string `md:"DisplayJson,required"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	strVal, _ := coerce.ToString(values["DisplayJson"])
	r.DisplayJson = strVal
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"DisplayJson": r.DisplayJson,
	}
}
