package sample

import "github.com/project-flogo/core/data/coerce"


type Input struct {
	Serial string `md:"serial,required"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	strVal, _ := coerce.ToString(values["serial"])
	r.Serial = strVal
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"serial": r.Serial,
	}
}




