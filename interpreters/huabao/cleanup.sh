#!/bin/bash

# Display what we're deleting
echo "Cleaning up unused files in the features directory..."

# These are the only directories we need to keep
REQUIRED_DIRS=(
  "features/huabao_protocol"
  "features/jono"
)

# These are files we need to keep for tests
REQUIRED_TEST_FILES=(
  "features/huabao_protocol/data.bin"
  "features/huabao_protocol/init_test.go"
  "features/jono/init_test.go"
)

# Delete any remaining backup files
find features -name "*copy*" -o -name "*.bak" -o -name "*~" -o -name "*.old" | while read file; do
  echo "Deleting backup file: $file"
  rm -f "$file"
done

echo "Cleanup complete!"
