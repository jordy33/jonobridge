#!/usr/bin/env python3
import paho.mqtt.client as mqtt
import time
import json
import argparse
import os

# Define the specific test message
DVR_TEST_MESSAGE = "$$dc0174,30,V114,0370703,,250613 091038,A0008,-99,9,-122688000,19,37,281100024,12.00,7800,0000000000009383,0000000000000000,0.00,0.00,0.00,-2040896705,0.00,0,0|0.00|0|0|0|0|0|0|2258,1#"

# Define the decoded information for comparison
EXTRACTED_VALUES = """
--- Extracted Values ---
Event: V114
IMEI: 0370703
Datetime: 250613 091038
Longitude Degrees: -99
Longitude Minutes: 9
Longitude Seconds: -122688000
Latitude Degrees: 19
Latitude Minutes: 37
Latitude Seconds: 281100024
Speed: 12.00
Heading: 7800
----------------------
"""

CONSTRUCTED_DATA = """
--- Constructed Data String ---
IMEI: 0370703
Type: DVR
Event: V114
Latitude: 19.624475
Longitude: -99.153408
Speed: 12.00
Azimuth: 78
Full data string: IMEI:0370703,DVR,V114,19.624475,-99.153408,12.00,78
-----------------------------
"""

# Record whether we received a response
received_response = False
received_data = None

def on_connect(client, userdata, flags, rc):
    """Callback when client connects to the MQTT broker"""
    if rc == 0:
        print(f"Connected to MQTT broker successfully")
        # Subscribe to the jonoprotocol topic to receive parsed results
        client.subscribe("tracker/jonoprotocol")
        print("Subscribed to tracker/jonoprotocol")
    else:
        print(f"Failed to connect to MQTT broker, return code {rc}")
        exit(1)

def on_message(client, userdata, msg):
    """Callback when a message is received"""
    global received_response, received_data
    try:
        payload = msg.payload.decode('utf-8')
        print(f"\nReceived response on topic {msg.topic}:")
        
        # Try to prettify JSON for better readability
        try:
            parsed_json = json.loads(payload)
            pretty_json = json.dumps(parsed_json, indent=2)
            print(pretty_json)
        except:
            # If JSON parsing fails, just print the raw payload
            print(payload)
        
        received_response = True
        received_data = payload
        
    except Exception as e:
        print(f"Error processing received message: {e}")

def main():
    parser = argparse.ArgumentParser(description='Send a specific DVR format message to test Huabao protocol parsing')
    parser.add_argument('--broker', '-b', default=os.environ.get('MQTT_BROKER_HOST', 'localhost'),
                      help='MQTT broker hostname or IP (default: MQTT_BROKER_HOST env var or localhost)')
    parser.add_argument('--topic', '-t', default='tracker/from-udp',
                      help='MQTT topic to publish to (default: tracker/from-udp)')
    parser.add_argument('--wait', '-w', type=float, default=10.0,
                      help='Time to wait for response in seconds (default: 10.0)')
    args = parser.parse_args()

    # Create MQTT client
    client_id = f"huabao_dvr_test_{int(time.time())}"
    client = mqtt.Client(protocol=mqtt.MQTTv311)
    client.on_connect = on_connect
    client.on_message = on_message
    
    # Connect to broker
    print(f"Connecting to MQTT broker at {args.broker}...")
    try:
        client.connect(args.broker, 1883, 60)
    except  Exception as e:
        print(f"Failed to connect to broker: {e}")
        exit(1)
    
    # Start the network loop in a background thread
    client.loop_start()
    
    # Wait a moment to ensure connection is established
    time.sleep(1)
    
    # Print the expected decoded values
    print("\n=== EXPECTED VALUES FOR COMPARISON ===")
    print(EXTRACTED_VALUES)
    print(CONSTRUCTED_DATA)
    
    # Send test message
    print("\n=== SENDING DVR TEST MESSAGE ===")
    print(f"Topic: {args.topic}")
    print(f"Message: {DVR_TEST_MESSAGE}")
    client.publish(args.topic, DVR_TEST_MESSAGE)
    print("\nMessage sent! Waiting for response...")
    
    # Wait for response
    timeout = time.time() + args.wait
    while time.time() < timeout and not received_response:
        time.sleep(0.1)
    
    if not received_response:
        print(f"\nNo response received within {args.wait} seconds")
    else:
        print("\n=== TEST COMPLETED ===")
        print("DVR message was processed by the system")
    
    # Clean up
    client.loop_stop()
    client.disconnect()
    print("Disconnected from MQTT broker")

if __name__ == "__main__":
    main()