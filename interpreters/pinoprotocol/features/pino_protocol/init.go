package pino_protocol

// TODO: IMPLEMENT MODULE INSTEAD MAIN
//

// import (
// 	"fmt"
// 	"net"
// 	"pinoprotocol/features/pino_protocol/usecases"
// 	"sync"
// )

// var imeiStore sync.Map

// func Initialize(conn net.Conn) (string, error) {
// 	defer conn.Close()
// 	// fmt.Printf("Conexión establecida con el dispositivo: %s\n", conn.RemoteAddr())
// 	clientAddr := conn.RemoteAddr().String()
// 	buffer := make([]byte, 1024)
// 	for {
// 		n, err := conn.Read(buffer)
// 		if err != nil {

// 			return "", fmt.Errorf("error leyendo datos: %s", err)
// 		}

// 		data := buffer[:n]
// 		fmt.Printf("Datos recibidos: %X\n", data)
// 		if len(data) < 1 || (data[0] != 0x7E && data[0] != 0x78) {
// 			return "", fmt.Errorf("invalid frame: first by doesnt 0x7E ni 0x78")
// 		}
// 		if data[0] == 0x7E {
// 			return usecases.ProcessFrame(data), nil
// 		} else if data[0] == 0x78 {
// 			switch {
// 			case usecases.IsLoginPacket(data):
// 				imei, err := usecases.ExtractIMEI(data)
// 				if err != nil {
// 					return "", fmt.Errorf("error extrayendo IMEI: %s", err)
// 				}

// 				imeiStore.Store(clientAddr, imei)
// 				fmt.Printf("IMEI almacenado para %s: %s\n", clientAddr, imei)

// 				response := usecases.BuildLoginResponse(data)
// 				fmt.Printf("Enviando respuesta de Login: %X\n", response)
// 				conn.Write(response)
// 			case usecases.IsHeartbeatPacket(data):
// 				response := usecases.BuildHeartbeatResponse()
// 				fmt.Printf("Enviando respuesta de Heartbeat: %X\n", response)
// 				conn.Write(response)
// 			case usecases.IsStandardLocationPacket(data):
// 				fmt.Println("Procesando paquete de ubicación estándar...")
// 				imeiValue, ok := imeiStore.Load(clientAddr)
// 				if !ok {
// 					return "", fmt.Errorf("error IMEI unknown to %s", clientAddr)
// 				}
// 				imei, ok := imeiValue.(string)
// 				if !ok {
// 					return "", fmt.Errorf("error IMEI for %s is not a string", clientAddr)
// 				}

// 				data, err := usecases.DecodeStandardLocationData(data, imei)
// 				if err != nil {
// 					return "", fmt.Errorf("error decoding location data, %s", err)

// 				}
// 				jsonData, err := data.ToJSON()
// 				if err != nil {
// 					return "", fmt.Errorf("error decoding json location data, %s", err)

// 				}
// 				return string(jsonData), nil
// 			default:
// 				return "", fmt.Errorf("packet unknown")
// 			}
// 		}

// 	}
// }
