package jsontools

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Int64String int64
type float64String float64

func (i Int64String) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%d", i))
}

func (f float64String) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.2f", f))
}

type ResponseID struct {
	ID Int64String `json:"id"`
}

func WriteInt64ID(w http.ResponseWriter, id int64) {
	json.NewEncoder(w).Encode(Int64String(id))
}

func WtiteJSON(w http.ResponseWriter, val any) {
	json.NewEncoder(w).Encode(val)
}
