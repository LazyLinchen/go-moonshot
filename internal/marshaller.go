package go_moonshot

import "encoding/json"

type Marshaller interface {
	Marshal(value any) ([]byte, error)
}

type JSONMarshaller struct{}

func (hm *JSONMarshaller) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}
