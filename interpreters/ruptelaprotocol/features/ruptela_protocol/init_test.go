package ruptela_protocol

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInicializater(t *testing.T) {
	data := "4e9caf2c000007d608f11a1480ba00ed00000a00000b07030500ad001b18011d666b03410491d08996000056c272013a3e9700"

	result, err := Initialize(data)
	fmt.Print(result)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestInitializaWithBinaryData(t *testing.T) {
	filePath := "data_ruptela.bin"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("No se pudo leer el archivo %s: %v", filePath, err)
	}
	dataString := string(data)

	result, err := Initialize(dataString)
	assert.NoError(t, err, "La función devolvió un error inesperado")

	expectedData := result
	assert.Equal(t, expectedData, result)
}
