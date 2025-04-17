package jsontools

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
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
	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(ResponseID{
		ID: Int64String(id),
	})
}

func WtiteJSON(w http.ResponseWriter, val any) {
	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(val)
}
