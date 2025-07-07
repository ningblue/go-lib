package gorm

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func ConvertStructpbToJSON(pb interface{}) datatypes.JSON {
	data, err := json.Marshal(pb)
	if err != nil {
		return nil
	}
	return datatypes.JSON(data)
}
