package jono

import (
	"fmt"
	"ruptelaprotocol/features/jono/usecases"
)

func Initialize(data string) (string, error) {
	parsedData, err := usecases.GetDataJono(data)
	if err != nil {

		return "", fmt.Errorf("error parsed data %s", err)
	}
	return parsedData, nil

}
