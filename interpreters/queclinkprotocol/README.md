# Queclink Protocol Parser in Go

This Go code implements a protocol parser for Queclink GPS tracking devices. It processes messages from these devices, parsing them into a map of key-value pairs based on the device type and message type. The supported devices include GV300W, GL320M, and GV350M.

## Overview

The code is designed to handle various message formats from Queclink devices, extracting relevant data such as device type, IMEI, and location information. It integrates with a MySQL database to fetch missing coordinates and uses an external `Mapas` package for field mappings and event codes.

---

## Main Components

### 1. `Message_Format` Function

```go
func Message_Format(passline []byte, port string, logs bool) map[string]string