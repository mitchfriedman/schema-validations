package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

type errResponse struct {
	Errors []string `json:"errors"`
}

func main() {
	schema, err := loadSchema()
	if err != nil {
		panic(fmt.Sprintf("failed to load schema.json: %v", err))
	}

	handler := validate(schema, process)
	http.ListenAndServe(":8000", handler)
}

func process(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("valid request"))
}

func validate(schema *gojsonschema.Schema, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		requestJSON := gojsonschema.NewBytesLoader(body)
		result, err := schema.Validate(requestJSON)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !result.Valid() {
			if err := writeError(result.Errors(), w); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func writeError(errors []gojsonschema.ResultError, w http.ResponseWriter) error {
	var r errResponse

	for _, e := range errors {
		r.Errors = append(r.Errors, e.String())
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	w.Write(b)
	w.WriteHeader(http.StatusBadRequest)

	return nil
}

func loadSchema() (*gojsonschema.Schema, error) {
	loader := gojsonschema.NewStringLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, err
	}

	return schema, nil
}
