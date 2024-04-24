package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var url = "https://go.getblock.io/..."

func GetEntity[T any](request Request) (*T, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var entity T
	err = json.NewDecoder(response.Body).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return &entity, nil
}
