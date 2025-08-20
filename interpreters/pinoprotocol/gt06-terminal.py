import binascii

def calculate_crc16(data):
    """
    Calculate CRC-16 (CRC-ITU) for the given data.
    Uses the lookup table from the GT06 protocol documentation appendix.
    """
    # CRC-ITU lookup table from the documentation
    crctab16 = [
        0x0000, 0x1189, 0x2312, 0x329B, 0x4624, 0x57AD, 0x6536, 0x74BF,
        0x8C48, 0x9DC1, 0xAF5A, 0xBED3, 0xCA6C, 0xDBE5, 0xE97E, 0xF8F7,
        0x1081, 0x0108, 0x3393, 0x221A, 0x56A5, 0x472C, 0x75B7, 0x643E,
        0x9CC9, 0x8D40, 0xBFDB, 0xAE52, 0xDAED, 0xCB64, 0xF9FF, 0xE876,
        0x2102, 0x308B, 0x0210, 0x1399, 0x6726, 0x76AF, 0x4434, 0x55BD,
        0xAD4A, 0xBCC3, 0x8E58, 0x9FD1, 0xEB6E, 0xFAE7, 0xC87C, 0xD9F5,
        0x3183, 0x200A, 0x1291, 0x0318, 0x77A7, 0x662E, 0x54B5, 0x453C,
        0xBDCB, 0xAC42, 0x9ED9, 0x8F50, 0xFBEF, 0xEA66, 0xD8FD, 0xC974,
        0x4204, 0x538D, 0x6116, 0x709F, 0x0420, 0x15A9, 0x2732, 0x36BB,
        0xCE4C, 0xDFC5, 0xED5E, 0xFCD7, 0x8868, 0x99E1, 0xAB7A, 0xBAF3,
        0x5285, 0x430C, 0x7197, 0x601E, 0x14A1, 0x0528, 0x37B3, 0x263A,
        0xDECD, 0xCF44, 0xFDDF, 0xEC56, 0x98E9, 0x8960, 0xBBFB, 0xAA72,
        0x6306, 0x728F, 0x4014, 0x519D, 0x2522, 0x34AB, 0x0630, 0x17B9,
        0xEF4E, 0xFEC7, 0xCC5C, 0xDDD5, 0xA96A, 0xB8E3, 0x8A78, 0x9BF1,
        0x7387, 0x620E, 0x5095, 0x411C, 0x35A3, 0x242A, 0x16B1, 0x0738,
        0xFFCF, 0xEE46, 0xDCDD, 0xCD54, 0xB9EB, 0xA862, 0x9AF9, 0x8B70,
        0x8408, 0x9581, 0xA71A, 0xB693, 0xC22C, 0xD3A5, 0xE13E, 0xF0B7,
        0x0840, 0x19C9, 0x2B52, 0x3ADB, 0x4E64, 0x5FED, 0x6D76, 0x7CFF,
        0x9489, 0x8500, 0xB79B, 0xA612, 0xD2AD, 0xC324, 0xF1BF, 0xE036,
        0x18C1, 0x0948, 0x3BD3, 0x2A5A, 0x5EE5, 0x4F6C, 0x7DF7, 0x6C7E,
        0xA50A, 0xB483, 0x8618, 0x9791, 0xE32E, 0xF2A7, 0xC03C, 0xD1B5,
        0x2942, 0x38CB, 0x0A50, 0x1BD9, 0x6F66, 0x7EEF, 0x4C74, 0x5DFD,
        0xB58B, 0xA402, 0x9699, 0x8710, 0xF3AF, 0xE226, 0xD0BD, 0xC134,
        0x39C3, 0x284A, 0x1AD1, 0x0B58, 0x7FE7, 0x6E6E, 0x5CF5, 0x4D7C,
        0xC60C, 0xD785, 0xE51E, 0xF497, 0x8028, 0x91A1, 0xA33A, 0xB2B3,
        0x4A44, 0x5BCD, 0x6956, 0x78DF, 0x0C60, 0x1DE9, 0x2F72, 0x3EFB,
        0xD68D, 0xC704, 0xF59F, 0xE416, 0x90A9, 0x8120, 0xB3BB, 0xA232,
        0x5AC5, 0x4B4C, 0x79D7, 0x685E, 0x1CE1, 0x0D68, 0x3FF3, 0x2E7A,
        0xE70E, 0xF687, 0xC41C, 0xD595, 0xA12A, 0xB0A3, 0x8238, 0x93B1,
        0x6B46, 0x7ACF, 0x4854, 0x59DD, 0x2D62, 0x3CEB, 0x0E70, 0x1FF9,
        0xF78F, 0xE606, 0xD49D, 0xC514, 0xB1AB, 0xA022, 0x92B9, 0x8330,
        0x7BC7, 0x6A4E, 0x58D5, 0x495C, 0x3DE3, 0x2C6A, 0x1EF1, 0x0F78
    ]

    # Initialize with 0xFFFF
    fcs = 0xFFFF

    # Calculate CRC for each byte in the input data
    for byte in data:
        fcs = (fcs >> 8) ^ crctab16[(fcs ^ byte) & 0xFF]

    # Perform bitwise NOT and ensure it's a 16-bit value
    crc_result = ~fcs & 0xFFFF

    # Return CRC as 2 bytes, big-endian
    return crc_result.to_bytes(2, byteorder='big')

def generate_gt06_command(command_str, serial_no_hex, language_code):
    """
    Generates a GT06 server-to-terminal command packet (Protocol 0x80).

    Args:
        command_str: The command content string (e.g., "WHERE#").
        serial_no_hex: The 2-byte information serial number as a hex string
                       (e.g., "05F2").
        language_code: The language code (1 for Chinese, 2 for English).

    Returns:
        The complete command packet as a hexadecimal string, or None if input is invalid.
    """
    try:
        # --- Fixed Values ---
        start_bit = b'\x78\x78'
        protocol_no = b'\x80'
        # Example Server Flag Bit (can be customized if needed)
        server_flag = b'\x00\x00\x00\x01'
        stop_bit = b'\x0d\x0a'

        # --- Process Inputs ---
        command_content = command_str.encode('ascii')
        if language_code == 1:
            language_bytes = b'\x00\x01' # Chinese
        elif language_code == 2:
            language_bytes = b'\x00\x02' # English
        else:
            print("Error: Invalid language code. Use 1 for Chinese or 2 for English.")
            return None

        # Ensure serial number is 2 bytes (4 hex chars)
        if len(serial_no_hex) != 4:
             print(f"Error: Serial number hex '{serial_no_hex}' must be 4 characters (2 bytes).")
             return None
        serial_no_bytes = binascii.unhexlify(serial_no_hex)
        if len(serial_no_bytes) != 2:
             print(f"Error: Serial number '{serial_no_hex}' must be 2 bytes.")
             return None

        # --- Calculate Lengths ---
        # Length of Command = Server Flag Bit length (4) + Command Content length
        len_command_val = 4 + len(command_content)
        if len_command_val > 255:
            print("Error: Command content too long.")
            return None
        len_command_byte = len_command_val.to_bytes(1, byteorder='big')

        # Information Content = Length of Command byte + Server Flag + Command Content + Language
        information_content = len_command_byte + server_flag + command_content + language_bytes

        # Packet Length = len(Proto No + Info Content + Serial No + Error Check)
        core_data_len = 1 + len(information_content) + 2 # Proto No(1) + Info Content + Serial No(2)
        packet_len_val = core_data_len + 2 # Add 2 for CRC bytes
        if packet_len_val > 255:
             print("Error: Calculated Packet Length exceeds 1 byte.")
             return None
        packet_len_byte = packet_len_val.to_bytes(1, byteorder='big')

        # --- Calculate CRC ---
        # Data for CRC: Packet Length byte + Protocol No + Information Content + Serial No bytes
        data_for_crc = packet_len_byte + protocol_no + information_content + serial_no_bytes
        # Use the function based on the manual's table
        error_check = calculate_crc16(data_for_crc)

        # --- Assemble Final Packet ---
        full_packet = (
            start_bit +
            packet_len_byte +
            protocol_no +
            information_content +
            serial_no_bytes +
            error_check +
            stop_bit
        )

        return binascii.hexlify(full_packet).decode('ascii')

    except binascii.Error:
        print(f"Error: Invalid hexadecimal string for serial number: '{serial_no_hex}'")
        return None
    except Exception as e:
        print(f"An error occurred: {e}")
        return None

# --- Main Program ---
if __name__ == "__main__":
    print("GT06 Command Packet Generator (Server to Terminal - 0x80)")

    # Get user input
    cmd_input = input("Enter command string (e.g., WHERE#): ")
    sn_input = input("Enter 2-byte serial number as HEX string (e.g., 05F2): ")
    lang_input_str = input("Enter language (1 for Chinese, 2 for English): ")

    try:
        lang_input = int(lang_input_str)
        # Generate command
        generated_command = generate_gt06_command(cmd_input, sn_input, lang_input)

        if generated_command:
            print("\nGenerated Command Packet:")
            print(generated_command)

            # Verify against the example provided by the user
            example_cmd = "WHERE#"
            example_sn = "05F2" # Adjusted based on example packet dissection
            example_lang = 1
            example_packet = "787812800a00000001574845524523000105f2a2220d0a"

            if cmd_input == example_cmd and sn_input.upper() == example_sn.upper() and lang_input == example_lang:
                 print(f"\nComparing with example:")
                 print(f"Example : {example_packet}")
                 if generated_command == example_packet:
                     print("Result matches the example!")
                 else:
                     print("Result DOES NOT match the example.")

    except ValueError:
        print("Error: Invalid input for language. Please enter 1 or 2.")
    except KeyboardInterrupt:
        print("\nOperation cancelled by user.")