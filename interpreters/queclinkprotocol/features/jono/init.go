package jono

import (
	"fmt"
	"queclinkprotocol/features/jono/usecases"
)

func Initialize(data string) (string, error) {
	parsedData, err := usecases.GetDataJono(data)
	if err != nil {
		fmt.Printf("Error al procesar datos: %v\n", err)
		return "", fmt.Errorf("error parsed data")
	}
	return parsedData, nil

}
