## 5.4. Heartbeat Packet (status information packet)

Heartbeat packet is a data packet to maintain the connection between the terminal and the server.

### 5.4.1. Terminal Sending Heartbeat Packet to Server

| Format               | Length (Byte) |
|----------------------|---------------|
| Start Bit            | 2             |
| Packet Length        | 1             |
| Protocol Number      | 1             |
| **Information Content** |             |
| Terminal Information | 1             |
| Status               | 1             |
| Voltage Level        | 1             |
| GSM Signal Strength  | 1             |
| Alarm/Language       | 2             |
| Serial Number        | 2             |
| Error Check          | 2             |
| Stop Bit             | 2             |