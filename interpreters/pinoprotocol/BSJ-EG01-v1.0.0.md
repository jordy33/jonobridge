# BSJ-EG01 Protocol And Data Format

## Revision Record

| Version | Function Description | Modifier | Change the time |
|---------|----------------------|----------|-----------------|
| V1.0.0  | Initial version      |          | 2024/1/23       |

---

## BSJ EG01 Protocol V1.0.0

### 1. Protocol Basis

#### 1.1. Communication Mode
The communication mode adopted in this protocol shall comply with the relevant provisions of **JT/T 794**. The communication protocol adopts **TCP**, with the platform as the server and the terminal as the client.

#### 1.2. Data Type
The data types used in protocol messages are shown below:

| Data type | Description and requirements                              |
|-----------|----------------------------------------------------------|
| BYTE      | Unsigned single-byte integer (byte, 8 bits)              |
| WORD      | Unsigned double-byte integer (word, 16 bits)             |
| DWORD     | Unsigned four-byte integer (Dword, 32 bits)              |
| BYTE[n]   | n bytes                                                 |
| BCD[n]    | 8421 code, n bytes                                      |
| STRING    | GBK encoding, if there is no data, leave it blank        |

#### 1.3. Transmission Rules
The protocol uses **big-endian** (network byte order) to pass words and Dwords. The agreement is as follows:
- **Byte (BYTE)**: Transmitted in the form of a byte stream.
- **Word (WORD)**: First transmit the high eight bits, then the low eight bits.
- **Double word (DWORD)**: First transmit the high 24 bits, then the high 16 bits, then the high eight bits, and finally the low eight bits.

#### 1.4. Message Composition

##### 1.4.1 Message Composition
Each message is composed of a flag bit header, message header, message body, and check code. The message structure is shown below:

| Identification bit | Message header | Message body | Check code | Identification bit |
|--------------------|----------------|--------------|------------|--------------------|

##### 1.4.2 Flag Bit
Represented by `0x7e`. If `0x7e` appears in the check code, message header, or message body, escaping processing is required. The escaping rules are:
- `0x7e` ↔ `0x7d` followed by `0x02`
- `0x7d` ↔ `0x7d` followed by `0x01`

**Escaping Process:**
- **Sending a message**: Message encapsulation → Compute and fill in the check code → Escape.
- **Receiving a message**: Transfer and restore → Verify the check code → Parse the message.

**Example:**
Sending a data packet with content `0x30 0x7e 0x08 0x7d 0x55` is encapsulated as:  
`0x7e 0x30 0x7d 0x02 0x08 0x7d 0x01 0x55 0x7e`.

##### 1.4.3 Message Header
The message header content is detailed below:

| Starting byte | Field                  | Data type | Description                                                                 |
|---------------|------------------------|-----------|-----------------------------------------------------------------------------|
| 0             | Message ID             | WORD      |                                                                             |
| 2             | Message body properties| WORD      | The message body attribute format structure is shown below                  |
| 4             | Terminal mobile number | BCD[6]    | Converted according to the mobile number of the installed terminal. If less than 12 digits, add numbers in front (0 for mainland, area code for HK/Macao/TW) |
| 10            | Message serial number  | WORD      | Accumulated in a loop from 0 according to the sending order                 |
| 12            | Message packet encapsulation item |       | Present if the message body attribute indicates segmentation; otherwise absent |

**Message Body Attribute Format Structure:**

| 15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
|----|----|----|----|----|----|---|---|---|---|---|---|---|---|---|---|
| Reserved | Subpacket | Data encryption method | Message body length                     |

- **Data Encryption Method**:  
  - Bits 10-12 are encryption identification bits.  
  - All 0s: Message body is not encrypted.  
  - Bit 10 = 1: Message body encrypted by RSA algorithm.  
  - Other values reserved.

- **Subpacket**:  
  - Bit 13 = 1: Message body is a long message sent in packets; packet info in the encapsulation item.  
  - Bit 13 = 0: No packet encapsulation in the message header.

**Message Packet Encapsulation Item:**

| Starting byte | Field                  | Data Type | Description and requirements                   |
|---------------|------------------------|-----------|------------------------------------------------|
| 0             | Total number of packets| WORD      | Total number of packets after segmentation     |
| 2             | Packet serial number   | WORD      | Starts from 1                                   |

##### 1.4.4 Check Code
The check code is the **exclusive OR (XOR)** from the message header to the byte before the check code, occupying **1 byte**.

---

## Data Format

### 1.1 Terminal General Response [0x0001]
**Message ID**: `0x0001`

| Starting byte | Field            | Data type | Description and requirements                        |
|---------------|------------------|-----------|----------------------------------------------------|
| 0             | Response serial number | WORD | Corresponding to the serial number of the platform message |
| 2             | Reply ID         | WORD      | Corresponding to the ID of the platform message     |
| 4             | Result           | BYTE      | 0: Success/Confirmation; 1: Failure; 2: Message error; 3: Not supported |

### 1.2 Platform General Response [0x8001]
**Message ID**: `0x8001`

| Starting byte | Field            | Data type | Description and requirements                        |
|---------------|------------------|-----------|----------------------------------------------------|
| 0             | Response serial number | WORD | Corresponding to the serial number of the terminal message |
| 2             | Reply ID         | WORD      | Corresponding to the ID of the terminal message     |
| 4             | Result           | BYTE      | 0: Success/Confirmation; 1: Failure; 2: Message Error; 3: Not Supported; 4: Alarm Processing Confirmation |

### 1.1 Terminal Heartbeat [0x0002]
**Message ID**: `0x0002`  
The terminal heartbeat message body is **empty**.

| Starting byte | Field          | Data type | Description and requirements |
|---------------|----------------|-----------|-----------------------------|
| 0             | Battery power  | BYTE      | Energy percentage           |
| 1             | CSQ            | BYTE      |                             |
| 2             | Status         | BYTE      | 0: Operating mode; 1: Standby; 2: Turn off |

### 1.2 Terminal Registration [0x0100]
**Message ID**: `0x0100`

| Starting byte | Field          | Data type | Description and requirements                        |
|---------------|----------------|-----------|----------------------------------------------------|
| 0             | Province ID    | WORD      | Province where the vehicle is located (GB/T 2260, first 2 digits) |
| 2             | City/County ID | WORD      | City/County where the vehicle is located (GB/T 2260, last 4 digits) |
| 4             | Manufacturer ID| BYTE[5]   | 5 bytes, terminal manufacturer number              |
| 9             | Terminal model | BYTE[8]   | 8 bytes, manufacturer-defined, space-filled if < 8 |
| 17            | Terminal ID    | BYTE[7]   | 7 bytes, uppercase letters/numbers, 0x00-filled if < 7 |
| 24            | License plate color | BYTE | Per JT/T 415-2006 5.4.12; 0 if not plated          |
| 25            | License plate  | STRING    | Vehicle license plate or VIN if color = 0           |

**Note**: Supplementary instructions require Terminal model to be 20 bytes (fill with `0x00` if insufficient).

### 1.3 Terminal Registration Response [0x8100]
**Message ID**: `0x8100`

| Starting byte | Field            | Data type | Description and requirements                        |
|---------------|------------------|-----------|----------------------------------------------------|
| 0             | Response serial number | WORD | Corresponding to terminal registration message serial number |
| 2             | Result           | BYTE      | 0: Success; 1: Vehicle registered; 2: No vehicle; 3: Terminal registered; 4: No vehicle |
| 3             | Authentication code | STRING | Only available after success                       |

### 1.4 Terminal Logout [0x0003]
**Message ID**: `0x0003`  
The terminal logout message body is **empty**.

### 1.5 Terminal Authentication [0x0102]
**Message ID**: `0x0102`

| Starting byte | Field            | Data type | Description and requirements                        |
|---------------|------------------|-----------|----------------------------------------------------|
| 0             | Authentication code | STRING | Terminal reconnects to report authentication code  |

### 1.6 Set Terminal Parameters [0x8103]
**Message ID**: `0x8103`

| Starting byte | Field                  | Data type | Description and requirements |
|---------------|------------------------|-----------|-----------------------------|
| 0             | Total number of parameters | BYTE  |                             |
| 1             | Package Parameter Quantity |       | Parameter Item Format (see below) |

**Parameter Item Format:**

| Field          | Data type | Description and requirements |
|----------------|-----------|-----------------------------|
| Parameter ID   | DWORD     | See Parameter Definitions   |
| Parameter length | BYTE    |                             |
| Parameter value| BYTE      |                             |

**Parameter Definitions:**

| Parameter ID | Data type | Description and requirements                        |
|--------------|-----------|----------------------------------------------------|
| 0x0001       | DWORD     | Terminal heartbeat sending interval (s)            |
| 0x0010       | STRING    | Main server APN (CDMA: PPP dial-up number)         |
| 0x0013       | STRING    | Main server address, IP or domain name             |
| 0x0018       | DWORD     | Server TCP port                                    |
| 0x0027       | DWORD     | Reporting interval during sleep (s, >0)            |
| 0x0029       | DWORD     | Default time reporting interval (s, >0)            |
| 0x0055       | DWORD     | Maximum speed (km/h)                               |
| 0x0056       | DWORD     | Overspeed duration (s)                             |
| 0x0080       | DWORD     | Vehicle odometer reading (1/10 km)                 |

### 1.7 Query Terminal Parameters [0x8104]
**Message ID**: `0x8104`  
The message body is **empty**. The terminal responds with `0x0104`.

### 1.8 Query Terminal Parameter Response [0x0104]
**Message ID**: `0x0104`

| Starting byte | Field                  | Data type | Description and requirements                        |
|---------------|------------------------|-----------|----------------------------------------------------|
| 0             | Response serial number | WORD      | Corresponding to the query message serial number   |
| 2             | Response Parameter Quantity | BYTE |                                              |
| 3             | Parameter list         |           | Parameter Item Format (see Parameter Definitions)  |

### 1.9 Terminal Control [0x8105]
**Message ID**: `0x8105`

| Starting byte | Field            | Data type | Description and requirements                        |
|---------------|------------------|-----------|----------------------------------------------------|
| 0             | Command word     | BYTE      | See Terminal Control Commands                      |
| 1             | Command parameters | STRING  | GBK-encoded, fields separated by ":"               |

**Terminal Control Commands:**

| Command word | Command parameters | Description and requirements |
|--------------|--------------------|-----------------------------|
| 4            | None               | Terminal reset              |

---

### 2.1 Position Information Report [0x0200]
The message body includes basic position information and an optional additional position information item list.

**Basic Position Information:**

| Starting byte | Field       | Data type | Description                                      |
|---------------|-------------|-----------|-------------------------------------------------|
| 0             | Alarm sign  | DWORD     | See Alarm Flag Bit Definition                   |
| 4             | Status      | DWORD     | See Status Bit Definition                       |
| 8             | Latitude    | DWORD     | Latitude × 10⁶ (degrees, 1/10⁶ precision)       |
| 12            | Longitude   | DWORD     | Longitude × 10⁶ (degrees, 1/10⁶ precision)      |
| 16            | Elevation   | WORD      | Altitude (m)                                    |
| 18            | Speed       | WORD      | Speed (1/10 km/h)                               |
| 20            | Direction   | WORD      | 0-359°, true north = 0, clockwise               |
| 21            | Time        | BCD[6]    | YY-MM-DD-hh-mm-ss (GMT+8)                       |

**Status Bit Definition:**

| Bit | Status                              |
|-----|-------------------------------------|
| 0   | 0: ACC off; 1: ACC on              |
| 1   | 0: Not positioned; 1: Positioned   |
| 2   | 0: North latitude; 1: South latitude |
| 3   | 0: East longitude; 1: West longitude |
| 4-9 |                                    |
| 10  |                                    |
| 11-31 |                                  |

**Alarm Flag Bit Definition:**

| Bit | Definition                        | Processing Instructions                   |
|-----|-----------------------------------|-------------------------------------------|
| 0   | SOS alarm                        |                                           |
| 1   | 1: Overspeed alarm               | Flag maintained until condition lifted    |
| 2-6 |                                  |                                           |
| 7   | 1: Terminal main power undervoltage | Flag maintained until condition lifted |
| 8-31|                                  |                                           |

**Additional Position Information Item Format:**

| Field                   | Data type | Description and requirements |
|-------------------------|-----------|-----------------------------|
| Additional information ID | BYTE    | 1–255                       |
| Additional information length | BYTE |                         |
| Additional Information  |           | See Additional Info Definitions |

**Additional Information Definitions:**

| Additional information ID | Length | Description and requirements                  |
|---------------------------|--------|----------------------------------------------|
| 0x01                     | 4      | Mileage, DWORD, 1/10 km (odometer reading)   |
| 0x30                     | 1      | Network signal strength                      |
| 0x31                     | 1      | GNSS positioning satellites (number of stars)|
| 0xEB                     |        | Extended data format (see Appendix D)        |

### 2.2 Location Information Query [0x8201]
**Message ID**: `0x8201`  
The message body is **empty**.

### 2.3 Location Information Query Response [0x0201]
**Message ID**: `0x0201`

| Starting byte | Field                  | Data type | Description and requirements                        |
|---------------|------------------------|-----------|----------------------------------------------------|
| 0             | Response serial number | WORD      | Corresponding to query message serial number       |
| 2             | Position Information Report |     | See Position Information Report (2.1)             |

### 2.4 Text Message Delivery [0x8300]
**Message ID**: `0x8300`

| Start byte | Field         | Data type | Description and requirements |
|------------|---------------|-----------|-----------------------------|
| 0          | Logo          | BYTE      | See Text Information Flag   |
| 1          | Text message  | STRING    |                             |

**Text Information Flag Meaning:**

| Bit | Logo            |
|-----|-----------------|
| 0   | 1: Emergency    |
| 1-7 |                 |

### 2.5 Report Text Message [0x6006]
**Message ID**: `0x6006`  
The terminal actively sends a text message; the platform must reply with a general response.

| Start byte | Field            | Data Type | Description and requirements           |
|------------|------------------|-----------|---------------------------------------|
| 0          | Text message encoding | BYTE | 0x00: BG2312; 0x01: UNICODE          |
| 1          | Text message     | STRING    |                                       |

---

## Appendix D: Uplink Extension Instructions

### Extended Instruction Format:

| Field      | Data Type | Description and requirements                  |
|------------|-----------|----------------------------------------------|
| Length     | WORD      | 2 bytes, includes instruction + data length  |
| Instruction| WORD      | 2 bytes                                      |
| Data       |           |                                              |

| Name                  | Length  | Instruction | Data Description                              |
|-----------------------|---------|-------------|----------------------------------------------|
| SIM ICCID number      | 0x000C  | 0x00B2      | 10-byte SIM card ICCID (Hex)                 |
| Extended alarm status bit | 0x0006 | 0x0089  | State[31-0], default 0xFFFFFFFF (see below)  |
| 4G base station       | 0x000B  | 0x00D8      | Country code (2B), Operator (1B), Area code (2B), Tower number (4B) |
| Extended alarm status bit | 0x0006 | 0x00C5  | State[31-0], default 0xFFFFFFFF (see below)  |
| External voltage value| 0x0006  | 0x002D      | Voltage (4B, Hex, e.g., 36B0 = 14000 mV)     |
| Percentage voltage    | 0x0003  | 0x00A8      | Voltage % (1B, Hex, e.g., 55 = 85%)          |
| Device IMEI number    | 0x0011  | 0x00D5      | 15-byte IMEI (Hex)                           |
| WiFi information      | n       | 0x00B9      | Hotspots (1B) + WIFI data (nB) (see below)   |

**Extended Alarm Status Bit (0x0089):**
- bit0: 1 = Battery off; 0 = Battery on
- bit1: 1 = Normal; 0 = Sleep
- bit4: 1 = Normal; 0 = Collision alarm
- bit8: 1 = Normal; 0 = Rapid acceleration
- bit9: 1 = Normal; 0 = Rapid deceleration
- bit12: 1 = Normal; 0 = Illegal removal
- bit25: 1 = Normal; 0 = Sharp turn alarm
- bit30: 1 = Normal; 0 = Pseudo base station detected
- bit31: 1 = Normal; 0 = Pseudo base station alarm

**Extended Alarm Status Bit (0x00C5):**
- bit3-bit4: [00] No positioning; [10] GPS positioning
- bit6: 1 = Normal; 0 = Vibration alarm
- bit14: 1 = Normal; 0 = Exposed to light alarm

**WiFi Information Format:**
- Number of hotspots: `0x01`–`0x05` (1B, Hex, max 5 groups)
- WIFI hotspot: `MAC address + signal value` (ASCII, e.g., `24:69:68:5d:2c:a5,-30`)

---

## Appendix E: 8300 Instruction Set

| Instruction                | Format Example                                      |
|----------------------------|----------------------------------------------------|
| Set family number (up to 5)| `<SPBSJ*P:BSJGPS*QQHM:17875175231,12342746346>`   |
| Set auto answer            | `<SPBSJ*P:BSJGPS*G0:1>`                           |
| Device calls back          | `<SPBSJ*P:BSJGPS*call:17875171231>`               |
| SOS alarm switch           | `<SPBSJ*P:BSJGPS*2:S:1>`                          |
| Configure IP               | `<SPBSJ*P:BSJGPS*T:047,107,222,141,7788*N:17811114444>` |
| Set positioning mode       | `<SPBSJ*P:BSJGPS*T:047,107,222,141,7788*N:17811114444>` |
| Set domain name            | `<SPBSJ*P:BSJGPS*Q:data,car900,com:7788>`         |

**Note**: Positioning modes:
- 0s: Standby mode
- 30s: Tracking mode
- 180s: Normal mode
- 3600s: Dotting mode