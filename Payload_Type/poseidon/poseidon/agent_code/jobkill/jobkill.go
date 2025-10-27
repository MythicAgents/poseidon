package jobkill

import "encoding/json"

type Arguments struct {
	ID string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["id"]; ok {
		e.ID = v.(string)
	}
	return nil
}
