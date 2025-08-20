package models

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CommandAAA CommandType = "AAA"
	CommandCCE CommandType = "CCE"
	CommandCCC CommandType = "CCC"
	CommandCFF CommandType = "CFF" // Added new CFF protocol type
)

type StartSignal string

const (
	StartSignalToServer StartSignal = "$$"
	StartSignalToDevice StartSignal = "@@"
)

type GeneralModel struct {
	StartSignal StartSignal
	Identifier  string
	DataLength  string
	IMEI        string
	CommandType CommandType
	Rest        string
	Message     string
}

func ParseGeneralFields(log string) (GeneralModel, error) {
	//fmt.Println(log)
	if len(log) < 25 {
		return GeneralModel{}, fmt.Errorf("error data too short: %s", log)
	}

	parts := strings.Split(log, ",")
	if len(parts) < 3 {
		return GeneralModel{}, fmt.Errorf("invalid message format, insufficient parts: %s", log)
	}

	// Ensure the first part is at least 2 characters long
	if len(parts[0]) < 2 {
		return GeneralModel{}, fmt.Errorf("invalid start signal format: %s", log)
	}

	startSignal := StartSignal(parts[0][:2])
	if startSignal != StartSignalToDevice && startSignal != StartSignalToServer {
		return GeneralModel{}, fmt.Errorf("Invalid start signal: %s", log)
	}

	// Ensure we have enough characters to extract the identifier
	identifier := ""
	if len(parts[0]) >= 3 {
		identifier = parts[0][2:3]
	}

	// Ensure we have enough characters to extract the data length
	dataLength := ""
	if len(parts[0]) >= 4 {
		dataLength = parts[0][3:]
	}

	return GeneralModel{
		StartSignal: startSignal,
		Identifier:  identifier,
		DataLength:  dataLength,
		IMEI:        parts[1],
		CommandType: CommandType(parts[2]),
		Rest:        strings.Join(parts[3:], ","),
		Message:     log,
	}, nil
}
