# GT06 Protocol Examples

This file contains examples of GT06 protocol messages and their decoded contents, which can help with troubleshooting and development.

## Standard Location Packet Example

Raw packet: `78781F120B081D112E10CC027AC7EB0C46584900148F01CC00287D001FB8000380810D0A`

### Breakdown:
- Start Bit: `7878`
- Packet Length: `1F` (31 bytes)
- Protocol Number: `12` (GPS location data)
- DateTime: `0B081D112E10` = 2011-08-29 17:46:16
- GPS Info & Satellites: `CC` = 12 satellites with good signal
- Latitude: `027AC7EB` = 22.540356° N
- Longitude: `0C465849` = 114.068942° E
- Speed: `00` = 0 km/h
- Course/Status: `148F` = Direction 143°, GPS positioned, North latitude, East longitude
- MCC: `01CC` = 460 (China)
- MNC: `00` = 0
- LAC: `287D` = 10365
- CellID: `001FB8` = 8120
- Serial Number: `0003`
- Checksum: `8081`
- Stop Bit: `0D0A`

## Common Issues with CellID

Some devices may report CellID as `000000` when:
1. The device can't determine the cell tower information
2. The device is in an area with poor cellular coverage
3. There's an issue with the device's cellular module

The CellID (3 bytes) follows the LAC (2 bytes), so if you're seeing all zeros, it's likely that:
- The packet is correctly formed but the device couldn't get the tower ID
- There might be a hardware issue with the GSM module

## Debugging Tips

When debugging location packets with missing or zero CellID:
1. Check if the device has good cellular signal
2. Try moving to an area with better cellular coverage
3. Verify that the SIM card is properly inserted and activated
4. Check the packet bytes directly to ensure the CellID position contains zeros
