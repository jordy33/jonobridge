# BSJ-EG01 Protocol And Data Format

## Revision Record [cite: 1]

| Version | Function Description | Modifier | Change the time |
| :------ | :------------------- | :------- | :-------------- |
| V1.0.0  | Initial version      |          | 2024/1/23       |
[cite: 2]

---

# BSJ EG01 Protocol V1.0.0

## 1. Protocol basis

### 1.1. Communication Mode [cite: 3]

The communication mode adopted in this protocol shall comply with the relevant provisions of JT/T 794. The communication protocol adopts TCP, with the platform as the server and the terminal as the client[cite: 3].

### 1.2. Data Type [cite: 4]

The data types used in protocol messages are shown in Figure 1[cite: 4]:

**Figure 1 Data Types** [cite: 5]

| Data type | Description and requirements                                   |
| :-------- | :------------------------------------------------------------- |
| BYTE      | Unsigned single-byte integer (byte, 8 bits)                    |
| WORD      | Unsigned double-byte integer (word, 16 bits)                   |
| DWORD     | Unsigned four-byte integer (Dword, 32 bits)                    |
| BYTE[n]   | n bytes                                                        |
| BCD[n]    | 8421 code, n bytes                                             |
| STRING    | GBK encoding, if there is no data, leave it blank              |
[cite: 5]

### 1.3. Transmission rules [cite: 6]

The protocol uses big-endian network byte order to pass words and Dwords[cite: 6].

The agreement is as follows[cite: 7]:

* **Byte (BYTE)** transmission agreement: transmitted in the form of a byte stream[cite: 7].
* **Word (WORD)** transmission agreement: first transmit the high eight bits, then transmit the low eight bits[cite: 8].
* **Double word (DWORD)** transmission agreement: first transmit the high 24 bits, then transmit the high 16 bits, then transmit the high eight bits, and finally transmit the low eight bits[cite: 9].

### 1.4 Message composition [cite: 10]

#### 1.4.1 Message Composition [cite: 11]

Each message is composed of a flag bit header, message header, message body, and check code[cite: 11]. The message structure is shown in Figure 1[cite: 12]:

**Figure 1 Message structure diagram** [cite: 13]

| Identification bit | Message header | Message body | Check code | Identification bit |
| :----------------- | :------------- | :----------- | :--------- | :----------------- |
[cite: 13]

#### 1.4.2 Flag Bit [cite: 14]

Uses `$0\times7e$`. If `$0\times7e$` appears in the check code, message header, and message body, escaping processing is required[cite: 14].

The escaping rules are defined as follows[cite: 15]:

* `$0\times7e$` -> `$0\times7d$` followed by `$0\times02$` [cite: 15]
* `$0\times7d$` -> `$0\times7d$` followed by `$0\times01$` [cite: 16]

**Escaping process:**

* **Sending a message:** message encapsulation -> compute and fill in the check code -> escape[cite: 17].
* **Receiving a message:** transfer and restore -> verify the check code -> parse the message[cite: 17].

**Example:**
Sending a data packet with the content of `$0\times30$ $0\times7e$ $0\times08$ $0\times7d$ $0\times55$` is encapsulated as follows: `$0\times7e$ $0\times30$ $0\times7d$ $0\times02$ $0\times08$ $0\times7d$ $0\times01$ $0\times55$ $0\times7e$`[cite: 17].

#### 1.4.3 Message header [cite: 18]

The message header content is detailed in Figure 2[cite: 18].

**Figure 2 Message header content** [cite: 19]

| Starting byte | Field                   | Data type | Description                                                                                                                                                                                                |
| :------------ | :---------------------- | :-------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0             | Message ID              | WORD      |                                                                                                                                                                                                            |
| 2             | Message body properties | WORD      | The message body attribute format structure is shown in Figure 2                                                                                                                                           |
| 4             | Terminal mobile phone number | BCD[6]    | Converted according to the mobile number of the installed terminal. If the mobile number is less than 12 digits, add numbers in front. Mainland mobile numbers add the number 0, Hong Kong, Macao and Taiwan add numbers according to their area code |
| 10            | Message serial number   | WORD      | Accumulated in a loop from 0 according to the sending order                                                                                                                                              |
[cite: 19]

**Figure 2 Message body attribute format structure diagram** [cite: 21]

| 15       | 14 | 13       | 12                     | 11 | 10                     | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0                   |
| :------- | :- | :------- | :--------------------- | :- | :--------------------- | :- | :- | :- | :- | :- | :- | :- | :- | :- | :------------------ |
| Reserved |    | Subpacket | Data encryption method |    | Data encryption method |   |   |   |   |   |   |   |   |   | Message body length |

* **Message packet encapsulation item**: If the relevant flag bit in the message body attribute determines the message segmentation processing, this item has content, otherwise this item does not exist[cite: 21].

**Data encryption method:** [cite: 22]

* Bit10-bit12 are data encryption identification bits[cite: 22].
* When these three bits are all 0, it means the message body is not encrypted[cite: 22].
* When the 10th bit is 1, it means that the message body is encrypted by RSA algorithm[cite: 23].
* Other reservations[cite: 24].

**Subpacket:** [cite: 25]

* If the 13th bit in the message body attribute is 1, it means that the message body is a long message and will be sent in packets[cite: 25].
* The specific packetization information is determined by the message packet encapsulation item[cite: 25].
* If the 13th bit is 0, there is no message packet encapsulation in the message header[cite: 26].

**Figure 3 Contents of message packet encapsulation items** [cite: 27]

| Starting byte | Field                    | Data Type | Description and requirements                                        |
| :------------ | :----------------------- | :-------- | :------------------------------------------------------------------ |
| 0             | Total number of message packets | WORD      | The total number of packets after the message is subpacketed        |
| 2             | Packet serial number     | WORD      | Start from 1                                                        |
[cite: 28]

#### 1.4.4 Check code [cite: 29]

The check code refers to the exclusive OR from the message header to the byte before the check code, occupying 1 byte[cite: 29].

## Data Format

### 1.1. Terminal general response 【0001】 [cite: 31]

* **Message ID:** `$0\times0001$` [cite: 31]
* The message body data format is shown in Figure 4[cite: 31].

**Figure 4 Terminal general response message body data format** [cite: 32]

| Starting byte | Field               | Data type | Description and requirements                       |
| :------------ | :------------------ | :-------- | :------------------------------------------------- |
| 0             | Response serial number | WORD      | Corresponding to the serial number of the platform message |
| 2             | Reply ID            | WORD      | Corresponding to the ID of the platform message      |
| 4             | result              | BYTE      | 0: Success/confirmation; 1 Failure; 2 Message error; 3: Not supported |
[cite: 32]

### 1.2. Platform general response 【8001】 [cite: 33]

* **Message ID:** `$0\times8001$`[cite: 33].
* The platform general response message body data format is shown in Figure 5[cite: 33].

**Figure 5 Platform General Response Message Body Data Format** [cite: 34]

| Starting byte | Field               | Data type | Description and requirements                                        |
| :------------ | :------------------ | :-------- | :------------------------------------------------------------------ |
| 0             | Response serial number | WORD      | Corresponding to the serial number of the terminal message          |
| 2             | Reply ID            | WORD      | Corresponding to the ID of the terminal message                     |
| 4             | result              | BYTE      | 0: Success/Confirmation; 1: Failure; 2: Message Error; 3: Not Supported; 4: Alarm Processing Confirmation |
[cite: 34]

### 1.1. Terminal heartbeat 【0002】 [cite: 36]

* **Message ID:** `$0\times0002$` [cite: 36]
* The terminal heartbeat message body contains battery power, CSQ, and status[cite: 36, 37].

| Starting byte | Field        | Data type | Description and requirements             |
| :------------ | :----------- | :-------- | :--------------------------------------- |
| 0             | battery power| Byte      | Energy percentage                        |
| 1             | CSQ          | Byte      |                                          |
| 2             | Status       | BYTE      | 0: operating mode 1: standby 2: turn off |
[cite: 37]

### 1.2. Terminal registration 【0100】 [cite: 38]

* **Message ID:** `$0\times0100$` [cite: 38]
* The terminal registration message body data format is shown in Figure 6[cite: 38].

**Figure 6 Terminal registration message body data format** [cite: 39]

| Starting byte | Field             | Data type | Description and requirements                                                                                                                                                                                                                                                            |
| :------------ | :---------------- | :-------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 0             | Province ID       | WORD      | Indicates the province where the vehicle with the installed terminal is located, 0 is reserved, and the platform takes the default value. The province ID uses the first two digits of the six-digit administrative division code stipulated in GB/T 2260                                |
| 2             | City and county ID | WORD      | Indicates the city area where the vehicle where the terminal is installed is reserved. 0 is reserved and the default value is taken by the platform. The city and county ID uses the last four digits of the six digits of the administrative division code specified in GB/T 2260. |
| 4             | Manufacturer ID   | BYTE [5]  | Five bytes, terminal manufacturer number                                                                                                                                                                                                                                              |
| 9             | Terminal model    | BYTE [8/20] | Eight bytes initially, supplementary instructions require 20 bytes. This terminal model is defined by the manufacturer. If the number of digits is insufficient, fill in spaces or `$0\times00$` as per note[cite: 39].                                                            |
| 17            | Terminal ID       | BYTE [7]  | Seven bytes, consisting of uppercase letters and numbers. This terminal ID is defined by the manufacturer. If there are insufficient digits, `$0\times00$` will be added[cite: 40, 41].                                                                                                   |
| 24            | license plate color| BYTE      | License plate color, in accordance with the provisions of 5.4.12 in JT/T 415-2006, when the license plate is not plated, the value is 0[cite: 41].                                                                                                                                       |
| 25            | license plate     | STRING    | Motor vehicle license plate issued by the public security and traffic management department. (Note: The supplementary instructions require that when the license plate color is 0, this indicates the vehicle VIN number)[cite: 41].                                                   |
[cite: 39, 40, 41]

### 1.3. Terminal registration response [8100] [cite: 42]

* **Message ID:** `$0\times8100$` [cite: 42]
* The terminal registration response message body data format is shown in Figure 7[cite: 42, 44].

**Figure 7 Terminal registration response message body data format** [cite: 43]

| Starting byte | Field               | Data type | Description and requirements                                                                                                          |
| :------------ | :------------------ | :-------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| 0             | Response serial number | WORD      | Corresponding to the serial number of the terminal registration message                                                               |
| 2             | result              | BYTE      | 0: Success; 1: Vehicle has been registered; 2: No such vehicle in the database; 3: Terminal has been registered; 4: No such vehicle in the database |
| 3             | Authentication code | STRING    | Only available after success                                                                                                        |
[cite: 43]

### 1.4. Terminal logout 【0003】 [cite: 44]

* **Message ID:** `$0\times0003$` [cite: 44]
* The terminal logout message body is empty[cite: 44].

### 1.5. Terminal authentication 【0102】 [cite: 45]

* **Message ID:** `$0\times0102$` [cite: 45]
* The terminal authentication message body data format is shown in Figure 8[cite: 45].

**Figure 8 Terminal authentication message body data format** [cite: 46]

| Starting byte | Field             | Data type | Description and requirements                           |
| :------------ | :---------------- | :-------- | :----------------------------------------------------- |
| 0             | Authentication code | STRING    | Terminal reconnects to report authentication code. |
[cite: 46]

### 1.6. Set terminal parameters 【8103】 [cite: 48]

* **Message ID:** `$0\times8103$` [cite: 48]
* The message body data format for setting terminal parameters is shown in Figure 9[cite: 48].

**Figure 9 Terminal parameter message body data format** [cite: 49]

| Starting byte | Field                   | Data type | Description and requirements                |
| :------------ | :---------------------- | :-------- | :------------------------------------------ |
| 0             | Total number of parameters | BYTE      |                                             |
| 1             | Package Parameter Quantity |           | Parameter Item Format (see Figure 10) [cite: 51] |
[cite: 49]

**Figure 10 Terminal parameter item data format** [cite: 50]

| Field           | Data type | Description and requirements                                                                                                                                    |
| :-------------- | :-------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Parameter ID    | DWORD     | Parameter ID Definition and Explanation (see Figure 11) [cite: 51]                                                                                                    |
| Parameter length | BYTE      |                                                                                                                                                                 |
| Parameter value |           | If it is a multi-valued parameter, multiple parameter items with the same ID are used in the message, such as the dispatch center phone number. [cite: 50] |
[cite: 50]

**Figure 11 Terminal Parameter Setting Each Parameter Item Definition and Explanation** [cite: 51]

| Parameter id | Data type | Description and requirements                                                                                                 |
| :----------- | :-------- | :--------------------------------------------------------------------------------------------------------------------------- |
| `$0\times0001$`  | DWORD     | Terminal heartbeat sending interval, unit is (s)                                                                             |
| `$0\times0010$`  | STRING    | Main server APN, wireless communication dial-up access point. If the network format is CDMA, this is the PPP dial-up number. |
| `$0\times0013$`  | STRING    | Main server address, IP or domain name                                                                                     |
| `$0\times0018$`  | DWORD     | Server TCP port                                                                                                              |
| `$0\times0027$`  | DWORD     | Reporting interval during sleep, unit is seconds (s), >0                                                                     |
| `$0\times0029$`  | DWORD     | Default time reporting interval, unit is seconds (s), >0                                                                     |
| `$0\times0055$`  | DWORD     | Maximum speed in kilometers per hour (km/h)                                                                                |
| `$0\times0056$`  | DWORD     | Overspeed duration, unit is seconds (s)                                                                                    |
| `$0\times0080$`  | DWORD     | Vehicle odometer reading, 1/10km                                                                                             |
[cite: 52]

### 1.7 Query terminal parameters 【8104】 [cite: 54]

* **Message ID:** `$0\times8104$` [cite: 54]
* Query Terminal Parameters Message Body is empty, terminal uses `$0\times0104$` instruction to respond[cite: 54].

### 1.8 Query terminal parameter response 【0104】 [cite: 54]

* **Message ID:** `$0\times0104$` [cite: 54]
* Query Terminal Parameters Response Message Body Data Format (see Figure 12)[cite: 54, 55].

**Figure 12 Query Terminal Parameters Response Message Body Data Format** [cite: 56]

| Starting byte | Field                     | Data type | Description and requirements                                                              |
| :------------ | :------------------------ | :-------- | :---------------------------------------------------------------------------------------- |
| 0             | Response serial number     | WORD      | Corresponding to the serial number of the terminal parameter query message              |
| 2             | Response Parameter Quantity | BYTE      |                                                                                           |
| 3             | Parameter list            |           | Parameter Item Format and Definition (see Figure 11) [cite: 51]                               |
[cite: 56]

### 1.9 Terminal control 【8105】 [cite: 57]

* **Message ID:** `$0\times8105$` [cite: 57]
* Terminal Control Message Body Data Format (see Figure 13)[cite: 57, 59].

**Figure 13 Terminal control message body data format** [cite: 58]

| Starting byte | Field             | Data type | Description and requirements                                                                                                                                  |
| :------------ | :---------------- | :-------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 0             | Command word      | BYTE      | Terminal Control Command Explanation (see Figure 14) [cite: 60]                                                                                                   |
| 1             | Command parameters | STRING    | The command parameter format is described later. Each field is separated by half-width "; ". Each STRING field is first processed by GBK encoding and then composed of a message. |
[cite: 58]

**Figure 14 Terminal control command description** [cite: 60]

| Command word | Command parameters | DESCRIPTION AND REQUIREMENTS |
| :----------- | :----------------- | :--------------------------- |
| 4            | none               | Terminal reset               |
[cite: 60]

### 2.1 Position Information Report【0200】 [cite: 60]

The position information report message body is composed of basic position information and additional position information item list[cite: 60]. The message structure is shown in Figure 3[cite: 61].

**Figure 3 Location report message structure diagram** [cite: 61]

| Basic location information | Additional Position Information Item List |
| :------------------------- | :---------------------------------------- |
[cite: 61]

* The additional position information item list is composed of various additional position information items, or it may not exist, determined by the length field in the message header[cite: 61].
* Basic Position Information Data Format (see Figure 16)[cite: 62].

**Figure 16 Basic location information data format** [cite: 63]

| Starting byte | Field      | Data type | Description                                                                              |
| :------------ | :--------- | :-------- | :--------------------------------------------------------------------------------------- |
| 0             | Alarm sign | DWORD     | Alarm Flag Bit Definition (see Figure 18) [cite: 66]                                        |
| 4             | Status     | DWORD     | Status Bit Definition (see Figure 17)                                                    |
| 8             | Latitude   | DWORD     | Latitude value multiplied by $10^6$ in degrees, accurate to one millionth of a degree[cite: 63]. |
| 12            | Longitude  | DWORD     | Longitude value multiplied by $10^6$ in degrees, accurate to one millionth of a degree[cite: 64].|
| 16            | Elevation  | WORD      | Altitude, unit is meters (m) [cite: 65]                                                  |
| 18            | Speed      | WORD      | 1/10km/h [cite: 65]                                                                      |
| 20            | Direction  | WORD      | 0-359 , true north is 0 , clockwise [cite: 65]                                           |
| 21            | Time       | BCD[6]    | YY-MM-DD-hh-mm-ss ( GMT+8 , all subsequent times in this standard use this time zone) [cite: 65] |
[cite: 63, 64, 65]

**Figure 17 status bit definition** [cite: 65]

| Bit   | Status                 |
| :---- | :--------------------- |
| 0     | 0: ACC off 1: ACC on |
| 1     | 0: Not positioned 1: Positioned |
| 2     | 0: North latitude 1: South latitude |
| 3     | 0: East longitude 1: West longitude |
| 4-9   | \_                     |
| 10    |                        |
| 11-31 | \_                     |
[cite: 65]

**Figure 18 Alarm standard bit definition** [cite: 66]

| Bit   | Definition                             | Processing Instructions                                                |
| :---- | :------------------------------------- | :--------------------------------------------------------------------- |
| 0     | SOS alarm                              |                                                                        |
| 1     | 1 : Overspeed alarm                    | The flag is maintained until the alarm condition is lifted[cite: 66]. |
| 2-6   | \_                                     |                                                                        |
| 7     | 1 : Terminal main power supply undervoltage | The flag is maintained until the alarm condition is lifted[cite: 67]. |
| 8-31  |                                        |                                                                        |
[cite: 66, 67]

**Figure 19 Position additional information item format** [cite: 66]

| Field                    | Type of data | Description and requirements                                     |
| :----------------------- | :----------- | :--------------------------------------------------------------- |
| Additional information ID | BYTE         | 1 ~ 255                                                          |
| Additional information length | BYTE         |                                                                  |
| Additional Information    |              | 20 for additional information definitions (See Figure 20) [cite: 69] |
[cite: 66]

**Figure 20 Additional information definitions** [cite: 69]

| Additional information ID | Additional information length | DESCRIPTION AND REQUIREMENTS                                                                        |
| :------------------------ | :---------------------------- | :-------------------------------------------------------------------------------------------------- |
| `$0\times01$`               | 4                             | Mileage, DWORD , 1/10km , corresponding to the vehicle odometer reading                            |
| `$0\times30$`               | 1                             | Network signal strength                                                                             |
| `$0\times31$`               | 1                             | GNSS positioning satellites (number of stars used)                                                |
| `$0\timesEB$`               |                               | extended data format, compatible with 2929 extended protocol, see extended additional D Figure for details customize[cite: 69]. |
[cite: 69]

### 2.2 Location information query 【8201】 [cite: 70]

* **Message ID:** `$0\times8201$` [cite: 70]
* Position Information Query Message Body is empty[cite: 70].

### 2.3 Location information query response [0201] [cite: 71]

* **Message ID:** `$0\times0201$` [cite: 71]
* Position Information Query Response Message Body Data Format (see Figure 24)[cite: 71].

**Figure 24 Location information query response message body data format** [cite: 71]

| Starting byte | Field                   | Data type | Description and requirements                                         |
| :------------ | :---------------------- | :-------- | :------------------------------------------------------------------- |
| 0             | Response serial number  | WORD      | Corresponding to the serial number of the position information query message[cite: 71]. |
| 2             | Position Information Report |           | Position Information Report (see Section 2.1) [cite: 72]                   |
[cite: 71, 72]

### 2.4 Text message delivery [8300] [cite: 73]

* **Message ID:** `$0\times8300$` [cite: 73]
* Text Information Issued Message Body Data Format (see Figure 26)[cite: 73].

**Figure 26 Text message delivery message body data format** [cite: 73]

| Start byte | Field       | Data type | Description and requirements                         |
| :--------- | :---------- | :-------- | :--------------------------------------------------- |
| 0          | logo        | BYTE      | Text Information Flag Bit Meaning (see Figure 27) |
| 1          | text message | STRING    | Text Information Flag Bit Meaning (see Figure 27) |
[cite: 73]

**Figure 27 Text information flag meaning** [cite: 73]

| Bit   | logo        |
| :---- | :---------- |
| 0     | 1: Emergency |
| 1-7   | \_          |
[cite: 73]

### 2.5 [Report text message] [6006] [cite: 73]

* **Message ID:** `$0\times6006$` [cite: 73]
* The terminal actively sends a text message, and the platform must reply with a platform general response after receiving it[cite: 74].
* The specific format is as follows[cite: 74]:

| Start byte | Field             | Data Type | Description and requirements                                  |
| :--------- | :---------------- | :-------- | :------------------------------------------------------------ |
| 0          | text message Encoding | BYTE      | `=0\times00` BG2312 encoding method <br> `=0\times01` UNICODE encoding method |
| 1          | text message      | STRING    |                                                               |
[cite: 74]

---

## Appendix D: Uplink Extension Instructions [cite: 75]

**Extension Instruction Format:** [cite: 75]

| Field       | Data Type | Description and requirements                             |
| :---------- | :-------- | :------------------------------------------------------- |
| length      | WORD      | 2 bytes, the length includes the instruction length plus the data length |
| instruction | WORD      | 2 bytes                                                  |
| data        |           |                                                          |
[cite: 75]

**Extended Data:** [cite: 75]

| Name                      | Length Occupied bytes | Instruction | Data Description                                                                                                                                                                                                                            |
| :------------------------ | :-------------------- | :---------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| SIM ICCID number          | N+2 (10+2=12)         | `$0\times00B2$`   | 10-byte SIM card ICCID number, represented by Hex[cite: 75].                                                                                                                                                                                 |
| Extended alarm status bit | N+2 (4+2=6)           | `$0\times0089$`   | State[31~0] The default is `$0xFFFFFFFF$`[cite: 76]. <br> bit0: 1 battery switch - off; 0 battery switch - on (battery) [cite: 76] <br> bit1: 1 Terminal normal state; 0 Terminal sleep state (sleep) [cite: 77] <br> bit4: 1 normal; 0 collision alarm [cite: 77] <br> bit8: 1 normal; 0 rapid acceleration alarm [cite: 78] <br> bit9: 1 normal; 0 rapid deceleration alarm [cite: 78] <br> bit12: 1 normal; 0 illegal removal [cite: 79] <br> bit25: 1 normal; 0 sharp turn alarm [cite: 79] <br> bit30: 1 normal; 0 pseudo base station detected [cite: 80] <br> bit31: 1 Normal; 0 pseudo base station alarm [cite: 80] |
| 4G base station           | N+2 (9+2=11)          | `$0\times00D8$`   | Country code: Occupies 2 bytes, HEX representation, e.g., `$0\times01CC$` represents 460[cite: 80]. <br> Operator number: Occupies 1 byte, HEX representation, e.g., `$0\times00$`[cite: 80]. <br> Area code: Occupies 2 bytes, HEX representation, high bit first, low bit after, e.g., `$0\times262C$`[cite: 80]. <br> Tower number: Occupies 4 bytes, HEX representation, high bit first, low bit after, e.g., `$0\times04BA0102$`[cite: 80]. <br> Note: If there is no base station information, fill in all 0 for the corresponding area code and tower number[cite: 80]. |
| Extended alarm status bit | N+2 (4+2=6)           | `$0\times00C5$`   | State[31~0] Default is `$0xFFFFFFFF$`[cite: 81]. <br> bit3~bit4: [00] No positioning; [10] GPS positioning [cite: 82] <br> bit6: 1 normal; 0 vibration alarm [cite: 82] <br> bit14: 1 normal; 0 alarm when exposed to light [cite: 83]      |
| External voltage value    | N+2 (4+2=6)           | `$0\times002D$`   | Voltage Value Occupies 4 bytes, HEX representation, e.g., `36B0 = 14000`(mV)[cite: 83].                                                                                                                                                     |
| Percentage voltage        | N+2 (1+2=3)           | `$0\times00A8$`   | Voltage percentage occupies one byte, represented by HEX, e.g., `55 = 85`(%)[cite: 83].                                                                                                                                                   |
| Device IMEI number        | N+2 (15+2=17)         | `$0\times00D5$`   | 15-byte device IMEI number, Hex representation[cite: 83].                                                                                                                                                                              |
| WiFi information          | n+2                   | `$0\times00B9$`   | Composed of the number of WIFI hotspots (1B) + WIFI hotspots (nB)[cite: 84]. <br> Number of hotspots: `$0\times01$`~ `$0\times05$`, up to 5 groups, represented by HEX[cite: 84]. <br> WIFI hotspot format: WIFI MAC address + signal value, represented by ASCALL code. E.g.: `24:69:68:5d:2c:a5,-30`[cite: 84, 85]. <br> Multiple group format: `24:69:68:5d:2c:a5,-30,2e:d0:5a:42:16:ad,-69,5c:0e:8b:8b:ca:50,-70,5c:0e:8b:8b:ca:52,-70,78:a1:06:6f:bb:fe,-70`[cite: 85]. |

---

## Appendix E 8300 instruction set [cite: 86]

| Function                               | Instruction                                                    |
| :------------------------------------- | :------------------------------------------------------------- |
| Set the family number (up to 5)        | `<SPBSJ*P:BSJGPS*QQHM:17875175231,12342746346>`                 |
| Set auto answer (can only be connected after setting) | `<SPBSJ*P:BSJGPS*60:1>`                                        |
| The device calls back                  | `<SPBSJ*P:BSJGPS*call:17875171231>`                            |
| SOS alarm switch                       | `<SPBSJ*P:BSJGPS*2S:1>`                                        |
| Configure IP                           | `<SPBSJ*P:BSJGPS*T:047.107.222.141,7788*N:17811114444>`          |
| Set the positioning mode               | standby：`<SPBSJ*P:BSJGPS*C:0>` <br> tracking：`<SPBSJ*P:BSJGPS*C:30>` <br> normal mode：`<SPBSJ*P:BSJGPS*C:180>` <br> dotting mode：`<SPBSJ*P:BSJGPS*C:3600>` (0=standby, 30=tracking, 180=normal, 3600=dotting) |
| Set domain name                        | `<SPBSJ*P:BSJGPS*Q:data.car900.com:7788>`                      |
[cite: 86]