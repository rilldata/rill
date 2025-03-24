#!/usr/bin/env bash

# Validates that *.sql migration files are in sequential order.
# This is useful for ensuring that there are not missing migrations on release branches.

# Usage: ./check_migrations.sh [path/to/migrations]
# Example: ./scripts/check_migrations.sh runtime/drivers/sqlite/migrations
# 

DEFAULT_MIGRATION_DIR="admin/database/postgres/migrations/"

# Path as input or default to migrations directory
MIGRATIONS_DIR=${1:-$DEFAULT_MIGRATION_DIR}

# Change to the specified directory
cd "$MIGRATIONS_DIR" || { echo "Directory not found: $MIGRATIONS_DIR"; exit 1; }

# Initialize a counter
expected=1

# Loop through all .sql files sorted by name
for file in $(ls *.sql 2>/dev/null | sort); do
    echo "$file"
    # Extract the number from the filename
    if [[ $file =~ ^([0-9]{4})\.sql$ ]]; then
        number=${BASH_REMATCH[1]}
        
        # Check if the number matches the expected sequence
        if (( 10#$number != expected )); then
            printf "File %s is out of order. Expected %04d.sql.\n" "$file" "$expected"
            exit 1
        fi
        
        # Increment the expected number
        ((expected++))
    else
        echo "File $file does not match the expected naming convention."
        exit 1
    fi
done

echo "All .sql files are in sequential order."
