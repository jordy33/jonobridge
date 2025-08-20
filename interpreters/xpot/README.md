# Environment Variables for XPOT

XPOT is a service that consumes a SpotX API to update device positions and forward them to a tracking server.

## Required Environment Variables

```bash
# Required: SpotX Feed URL for device tracking
export XPOT_FEED="https://api.findmespot.com/spot-main-web/consumer/rest-api/2.0/public/feed/0BkM9B2i01vF8eigoq3T1XO5HgMfQmfQa/message.xml"
```

## Optional Environment Variables

```bash
# Server Configuration
export XPOT_FORWARD_HOST="server1.gpscontrol.com.mx:8500"    # Default tracking server

# Database Configuration
export MYSQL_HOST="127.0.0.1"    # Database host
export MYSQL_PORT="3306"         # Database port
export MYSQL_USER="gpscontrol"   # Database user
export MYSQL_PASS="qazwsxedc"    # Database password
export MYSQL_DB="bridge"         # Database name

# Application Configuration
export XPOT_POLLING_TIME="30"    # Polling interval in seconds (default: 30)
```

## Command Line Flags

The following command line flags are available:

- `-v`: Enable verbose output (debug mode)

## Usage Examples

1. Basic run with default configuration:
```bash
./xpot
```

2. Run with verbose output:
```bash
./xpot -v
```

3. Run with custom configuration:
```bash
export MYSQL_HOST="custom.db.host"
export MYSQL_PORT="3307"
./xpot -v
```

## Database Tables

The application automatically creates and manages two tables:

1. `devices` - Stores device information:
   - `imei`: Device identifier
   - `plates`: Vehicle plates
   - `vin`: Vehicle identification number
   - `protocol`: Protocol identifier
   - `password`: Device password
   - `log`: Device log messages
   - `ff0`: Additional identifier

2. `spotx_message` - Stores SpotX messages:
   - `id`: Message identifier
   - `messengerId`: SpotX messenger ID
   - `messengerName`: Messenger name
   - `unixTime`: Message timestamp
   - `messageType`: Type of message
   - `latitude`: GPS latitude
   - `longitude`: GPS longitude
   - `altitude`: GPS altitude
   - Additional metadata fields

## Logging

- Error messages are always logged with timestamps
- Debug messages (with `-v` flag) show:
  - Database operations
  - API requests
  - Message processing
  - Server communications

## Error Handling

The application will:
1. Exit if required environment variables are missing
2. Log errors with timestamps
3. Continue processing other messages if one fails
4. Automatically create required database tables
