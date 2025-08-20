import binascii
import struct

def calculate_checksum(data):
    """
    Calculate the BSJ protocol checksum (XOR of all bytes)
    """
    checksum = 0
    for byte in data:
        checksum ^= byte
    return bytes([checksum])

def generate_bsj_command(command_str, terminal_phone, serial_number):
    """
    Generates a BSJ-EG01 text message command packet (Message ID: 0x8300).
    
    Args:
        command_str: The command content string (e.g., "<SPBSJ*P:BSJGPS*C:30>")
        terminal_phone: The terminal's phone number (BCD encoded)
        serial_number: Message serial number (0-65535)
    
    Returns:
        The complete command packet as a hexadecimal string
    """
    try:
        # Start flag bit
        flag_bit = b'\x7e'
        
        # Message ID (0x8300 for text message delivery)
        message_id = struct.pack('>H', 0x8300)
        
        # Message body (command string)
        message_body = b'\x00' + command_str.encode('gbk')  # 0x00 = normal flag (not emergency)
        body_length = len(message_body)
        
        # Message body properties
        # bits 0-9: message body length
        # bits 10-12: encryption (0 = not encrypted)
        # bit 13: subpacket flag (0 = no subpackets)
        body_props = struct.pack('>H', body_length)
        
        # Terminal phone number (BCD[6])
        if len(terminal_phone) != 12:
            # Pad to 12 digits
            terminal_phone = terminal_phone.zfill(12)
        phone_bytes = bytes.fromhex(''.join([terminal_phone[i:i+2] for i in range(0, len(terminal_phone), 2)]))
        
        # Message serial number
        serial_bytes = struct.pack('>H', serial_number)
        
        # Header (without check code)
        header = message_id + body_props + phone_bytes + serial_bytes
        
        # Full data for checksum calculation
        data_for_checksum = header + message_body
        
        # Calculate checksum
        check_code = calculate_checksum(data_for_checksum)
        
        # Construct full packet before escaping
        raw_packet = header + message_body + check_code
        
        # Apply escaping rules:
        # 0x7e -> 0x7d 0x02
        # 0x7d -> 0x7d 0x01
        escaped_packet = bytearray()
        for byte in raw_packet:
            if byte == 0x7e:
                escaped_packet.extend(b'\x7d\x02')
            elif byte == 0x7d:
                escaped_packet.extend(b'\x7d\x01')
            else:
                escaped_packet.append(byte)
        
        # Full packet with flags
        full_packet = flag_bit + bytes(escaped_packet) + flag_bit
        
        return binascii.hexlify(full_packet).decode('ascii')
        
    except Exception as e:
        print(f"An error occurred: {e}")
        return None

# --- Main Program ---
if __name__ == "__main__":
    print("BSJ-EG01 Command Packet Generator")
    
    # Command examples from Appendix E
    print("\nAvailable command examples:")
    print("1. Set tracking mode:       <SPBSJ*P:BSJGPS*C:30>")
    print("2. Set standby mode:        <SPBSJ*P:BSJGPS*C:0>")
    print("3. Set IP:                  <SPBSJ*P:BSJGPS*T:047.107.222.141,7788*N:17811114444>")
    print("4. Set domain name:         <SPBSJ*P:BSJGPS*Q:data.car900.com:7788>")
    print("5. Set family number:       <SPBSJ*P:BSJGPS*QQHM:17875175231,12342746346>")
    
    # Get user input
    cmd_input = input("\nEnter command string: ")
    phone_input = input("Enter terminal phone number (12 digits): ")
    serial_input = input("Enter message serial number (0-65535): ")
    
    try:
        serial_number = int(serial_input)
        # Generate command
        generated_command = generate_bsj_command(cmd_input, phone_input, serial_number)
        
        if generated_command:
            print("\nGenerated Command Packet:")
            print(generated_command)
            
    except ValueError:
        print("Error: Invalid input for serial number. Please enter a number between 0 and 65535.")
    except KeyboardInterrupt:
        print("\nOperation cancelled by user.")