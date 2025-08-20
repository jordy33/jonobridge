### Analysis of Meitrack Protocol Stream for Driver Leave DMS Alarm

Data:
```
24246f3630392c3836363831313036323637343632302c4343452c000000000400690017000505000600070d14001502090800000900000a00000b00001608001704001902001ae3044023000602e2dd290103a02716fa040c63892f0c510000000d73bd00001c00200000030e0c4e0114005a02197e4b02000049090405000000000000004b050101023447fe0019000505000600070d14001502090800000900000a00000b00001608001705001901001ae304407e000502e2dd290103a02716fa041563892f0c510000000d7cbd0000060e0c4e0114005a02197e4b020000fe3142020a3235303430393136353534395f4348315f453132365331305f305f444d5328444141292e6a7067000000000000000000000000000000000000000000000000004909040500000000000000fe7947010202010001023f4348315f3235303430393136353534365f3235303430393136353630395f453132365331305f305f315f315f414441535f444d5328444141292e61766d7367fe800801020303000403004b050101023447690017000505000600070d14001502090800000900000a00000b00001608001706001902001ae6044023000602e2dd290103a02716fa041663892f0c510000000d7dbd00001c00200000030e0c4e0114005a02197e4b02000049090405000000000000004b050101023447690017000505000600070d14001502090800000900000a00000b00001608001706001901001ae4044023000602e2dd290103a02716fa042063892f0c510000000d87bd00001c00200000030e0c4e0114005a02197e4b02000049090405000000000000004b0501010234472a41340d0a
```

I've analyzed the hex stream you provided and can confirm that it does contain a Driver Leave DMS Alarm.

Evidence in the stream:
This is a CCE protocol packet from device with IMEI 866811062674620

The key evidence is found in this section of the hex stream:

Breaking this down:
```

fe31 - This is the ID for "AdditionalAlertInfoADASDMS" as defined in the protocol
42 and 02 - Protocol version/parameters
0a - This is the critical alarm code
The rest decodes to a filename: 250409165549_CH1_E126S10_0_DMS(DAA).jpg
The alarm code 0a corresponds to "Driver absence" in the Meitrack protocol's second protocol alarm types (AdditionalAlarmTypesSecondProtocol)
```

### Additional supporting evidence:

This contains another DMS(DAA) related file, suggesting a driver monitoring event was captured.

Conclusion:
The stream contains a Driver Leave/Absence DMS Alarm with supporting image and message files. The system correctly identified when the driver was absent from their position and triggered the appropriate alert.



### Yawning:
```
2025-04-09T12:37:48 equivalence to 2025-04-09T18:37:48Z 
"ListPackets":{"packet_1":{"AdditionalAlertInfoADASDMS":{"AlarmProtocol":"02","AlarmType":"Yawning","PhotoName":"250409183748_CH1_E126S6_0_DMS(DYA).jpg"}
```

### Driver Leave:
```
2025-04-09T12:38:39 equivalence to 2025-04-09T18:38:39Z 
{"AdditionalAlertInfoADASDMS":{"AlarmProtocol":"02","AlarmType":"Driver absence","PhotoName":"250409183839_CH1_E126S10_0_DMS(DAA).jpg"
```

### Drowsiness Eyes closed

```
2025-04-09T12:39:09 equivalence to 2025-04-09T18:39:09Z 
2025-04-09T12:38:03 equivalence to 2025-04-09T18:38:03Z  


"ListPackets":{"packet_1":{"AdditionalAlertInfoADASDMS":{"AlarmProtocol":"02","AlarmType":"Drowsiness","PhotoName":"250409183803_CH1_E126S5_0_DMS(DFW).jpg"
```

### Calling
2025-04-09T13:16:12 equivalence to 2025-04-09T19:36:12Z 

"ListPackets":{"packet_1":{"AdditionalAlertInfoADASDMS":{"AlarmProtocol":"02","AlarmType":"Calling","PhotoName":"250409191612_CH1_E126S7_0_DMS(CALL).jpg"