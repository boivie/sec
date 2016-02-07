package app
import "encoding/json"

func SerializeJSON(value interface{}) []byte {
	out, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	if string(out) == "null" {
		panic("Tried to serialize a nil pointer.")
	}
	return out
}