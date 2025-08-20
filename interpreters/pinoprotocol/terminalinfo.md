# Common GT06 Protocol Commands

Based on the GT06 protocol standard, here are other commands besides `WHERE#` that are typically supported:

## Device Control Commands

- `RESET#` - Restart the device

- `POWEROFF#` - Turn off the device

- `BEGIN#` - Initialize/activate the device

- `FACTORY#` - Reset to factory settings

## Tracking Commands

- `TIMER#` - Set position reporting interval (e.g., `TIMER#30#` for every 30 seconds)

- `MOVING#` - Set movement alarm parameters

- `SPEED#` - Set speed alarm threshold (e.g., `SPEED#100#` for 100 km/h)

- `SLEEP#` - Set sleep mode parameters

## Communication Setup

- `APN#` - Configure APN settings (e.g., `APN#apn_name#username#password#`)

- `SERVER#` - Set server IP and port (e.g., `SERVER#192.168.1.1#8080#`)

- `HBT#` - Set heartbeat interval (e.g., `HBT#10#` for 10 minutes)

- `CENTER#` - Set center phone number

## Vehicle Control

- `RELAY#` - Control relay for cutting/connecting oil and electricity

- `RESTORE#` - Restore oil and electricity

## Other Functions

- `MONITOR#` - Enter voice monitoring mode

- `SMSLINK#` - Request map link via SMS

- `IMEI#` - Query device IMEI

You can use any of these commands with your script by replacing `WHERE#` with the desired command.

Note: The exact command support may vary between different GT06 device implementations. Check your specific device's documentation for supported commands and their exact formats.