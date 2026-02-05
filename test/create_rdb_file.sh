#!/bin/bash

# Create RDB file with header, metadata, and database sections
OUTPUT_FILE="custom.rdb"

# Remove existing file if present
rm -f "$OUTPUT_FILE"

echo "Creating RDB file: $OUTPUT_FILE"

# ============================================
# HEADER SECTION
# ============================================
# Magic string "REDIS" + version "0011"
echo -ne '\x52\x45\x44\x49\x53\x30\x30\x31\x31' >> "$OUTPUT_FILE"
echo "✓ Header: REDIS0011"

# ============================================
# METADATA SECTION
# ============================================
# Metadata subsection for "redis-ver" = "6.0.16"
echo -ne '\xFA' >> "$OUTPUT_FILE"  # Metadata marker
echo -ne '\x09\x72\x65\x64\x69\x73\x2D\x76\x65\x72' >> "$OUTPUT_FILE"  # "redis-ver" (9 chars)
echo -ne '\x06\x36\x2E\x30\x2E\x31\x36' >> "$OUTPUT_FILE"  # "6.0.16" (6 chars)
echo "✓ Metadata: redis-ver = 6.0.16"

# ============================================
# DATABASE SECTION
# ============================================
# Database selector: DB 0
echo -ne '\xFE' >> "$OUTPUT_FILE"  # Database marker
echo -ne '\x00' >> "$OUTPUT_FILE"  # Database index 0
echo "✓ Database: Selected DB 0"

# Hash table size information
echo -ne '\xFB' >> "$OUTPUT_FILE"  # Hash table size marker
echo -ne '\x03' >> "$OUTPUT_FILE"  # 3 keys total
echo -ne '\x02' >> "$OUTPUT_FILE"  # 2 keys with expiry
echo "✓ Hash table: 3 keys, 2 with expiry"

# ============================================
# KEY-VALUE 1: "foobar" = "bazqux" (no expiry)
# ============================================
echo -ne '\x00' >> "$OUTPUT_FILE"  # Value type: string
echo -ne '\x06\x66\x6F\x6F\x62\x61\x72' >> "$OUTPUT_FILE"  # Key: "foobar" (6 chars)
echo -ne '\x06\x62\x61\x7A\x71\x75\x78' >> "$OUTPUT_FILE"  # Value: "bazqux" (6 chars)
echo "✓ Key: foobar = bazqux (no expiry)"

# ============================================
# KEY-VALUE 2: "foo" = "bar" (expires in milliseconds)
# ============================================
echo -ne '\xFC' >> "$OUTPUT_FILE"  # Expiry in milliseconds marker
# Timestamp: 1713824559637 in little-endian (8 bytes)
echo -ne '\x15\x72\xE7\x07\x8F\x01\x00\x00' >> "$OUTPUT_FILE"
echo -ne '\x00' >> "$OUTPUT_FILE"  # Value type: string
echo -ne '\x03\x66\x6F\x6F' >> "$OUTPUT_FILE"  # Key: "foo" (3 chars)
echo -ne '\x03\x62\x61\x72' >> "$OUTPUT_FILE"  # Value: "bar" (3 chars)
echo "✓ Key: foo = bar (expires at 1713824559637 ms)"

# ============================================
# KEY-VALUE 3: "baz" = "qux" (expires in seconds)
# ============================================
echo -ne '\xFD' >> "$OUTPUT_FILE"  # Expiry in seconds marker
# Timestamp: 1714089298 in little-endian (4 bytes)
echo -ne '\x52\xED\x2A\x66' >> "$OUTPUT_FILE"
echo -ne '\x00' >> "$OUTPUT_FILE"  # Value type: string
echo -ne '\x03\x62\x61\x7A' >> "$OUTPUT_FILE"  # Key: "baz" (3 chars)
echo -ne '\x03\x71\x75\x78' >> "$OUTPUT_FILE"  # Value: "qux" (3 chars)
echo "✓ Key: baz = qux (expires at 1714089298 sec)"

# ============================================
# END OF FILE
# ============================================
echo -ne '\xFF' >> "$OUTPUT_FILE"  # EOF marker
echo -ne '\x00\x00\x00\x00\x00\x00\x00\x00' >> "$OUTPUT_FILE"  # CRC64 checksum (zeros for simplicity)
echo "✓ EOF marker + checksum"

echo ""
echo "✅ RDB file created successfully: $OUTPUT_FILE"
echo ""
echo "To verify:"
echo "  hexdump -C $OUTPUT_FILE"
echo "  redis-check-rdb $OUTPUT_FILE"
echo ""
echo "To load into Redis:"
echo "  redis-server --dir . --dbfilename $OUTPUT_FILE"

# Optional: Display hex dump
echo ""
echo "Hex dump of created file:"
hexdump -C "$OUTPUT_FILE"
