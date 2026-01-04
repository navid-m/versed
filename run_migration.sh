#!/bin/bash

echo "========================================="
echo "Post ID Migration: INTEGER -> UUID"
echo "========================================="
echo ""
read -p "Press Enter to continue or Ctrl+C to cancel..."

echo ""
echo "Creating backup..."
BACKUP_FILE="data.db.backup.$(date +%Y%m%d_%H%M%S)"
cp data.db "$BACKUP_FILE"
echo "✓ Backup created: $BACKUP_FILE"

echo ""
echo "Building migration tool..."
go build -o migrate_posts migrate_posts_to_uuid.go

if [ $? -ne 0 ]; then
    echo "✗ Failed to build migration tool"
    exit 1
fi

echo "✓ Migration tool built"

echo ""
echo "Running migration..."
./migrate_posts

if [ $? -ne 0 ]; then
    echo ""
    echo "✗ Migration failed!"
    echo "Restoring backup..."
    cp "$BACKUP_FILE" data.db
    echo "✓ Backup restored"
    exit 1
fi

echo ""
echo "Cleaning up..."
rm migrate_posts

echo ""
echo "========================================="
echo "✓ Migration completed successfully!"
echo "========================================="
echo ""
echo "Backup saved as: $BACKUP_FILE"
echo "You can now start the versed service"
echo ""
