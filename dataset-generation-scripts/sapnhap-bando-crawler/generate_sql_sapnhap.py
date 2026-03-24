import json
import os

# File paths
input_json_path = './donvi_tinhthanh.json'
output_sql_path = 'bando_co_dvch.sql'

# Configuration
BATCH_SIZE = 1000  # Number of rows per INSERT statement

def clean_string(s):
    """Remove quotes and extra spaces if present, to ensure clean SQL string insertion."""
    if not s:
        return None
    # Strip leading/trailing whitespace
    s = s.strip()
    # Remove single quotes inside string
    s = s.replace("'", "")
    return s

def sql_escape(val):
    """Escape string for SQL single quote literal (' -> '') and wrap in quotes."""
    if val is None:
        return "NULL"
    if isinstance(val, str):
        escaped = val.replace("'", "''").replace("\\", "\\\\")
        return f"'{escaped}'"
    return "NULL"

def generate_sql():
    # Load JSON data
    if not os.path.exists(input_json_path):
        print(f"Error: Input file not found at {input_json_path}")
        return

    with open(input_json_path, 'r', encoding='utf-8') as f:
        try:
            data = json.load(f)
        except json.JSONDecodeError:
            print(f"Error: Invalid JSON in file {input_json_path}")
            return

    # Prepare SQL content
    sql_lines = []
    sql_lines.append("-- SQL Script for bando_co_dvch")
    sql_lines.append("-- Generated from: " + input_json_path)
    sql_lines.append("-- Date: 2025-03-15")
    sql_lines.append("-- Using PostgreSQL Bulk INSERT for performance")
    sql_lines.append("")
    sql_lines.append("BEGIN;")
    sql_lines.append("")

    # Generate Bulk Insert Statements
    insert_count = 0
    batch_count = 0
    current_batch = []

    for item in data:
        ma = item.get('ma')
        ten = item.get('ten')
        magoc = item.get('magoc')
        malk = item.get('malk')
        truocsapnhap = item.get('truocsapnhap')
        
        # Validate required fields
        if not ma or not ten:
            print(f"Skipping item due to missing required fields: {item}")
            continue
        
        # Clean values
        clean_ten = clean_string(ten) if ten else None
        clean_truocsapnhap = clean_string(truocsapnhap) if truocsapnhap else None
        
        # Handle magoc: set to NULL if it's '0'
        s_magoc = sql_escape(magoc) if magoc and magoc != '0' else "NULL"
        
        # Escape Values
        s_ma = sql_escape(ma) if ma else "NULL"
        s_ten = sql_escape(clean_ten) if clean_ten else "NULL"
        s_malk = sql_escape(malk) if malk else "NULL"
        s_truocsapnhap = sql_escape(clean_truocsapnhap) if clean_truocsapnhap else "NULL"
        
        # Build VALUES tuple (only 5 columns: ma, ten, magoc, malk, truocsapnhap)
        values_tuple = f"({s_ma}, {s_ten}, {s_magoc}, {s_malk}, {s_truocsapnhap})"
        current_batch.append(values_tuple)
        insert_count += 1
        batch_count += 1
        
        # If batch is full, write INSERT statement
        if batch_count >= BATCH_SIZE:
            values_str = ",\n    ".join(current_batch)
            sql_lines.append(f"INSERT INTO bando_co_dvch (ma, ten, magoc, malk, truocsapnhap)")
            sql_lines.append(f"VALUES")
            sql_lines.append(f"    {values_str};")
            sql_lines.append("")
            current_batch = []
            batch_count = 0
            print(f"Processed {insert_count} rows...")

    # Write remaining rows in last batch
    if current_batch:
        values_str = ",\n    ".join(current_batch)
        sql_lines.append(f"INSERT INTO bando_co_dvch (ma, ten, magoc, malk, truocsapnhap)")
        sql_lines.append(f"VALUES")
        sql_lines.append(f"    {values_str};")
        sql_lines.append("")

    sql_lines.append("COMMIT;")
    sql_lines.append("")
    sql_lines.append(f"-- End of Script. Total rows: {insert_count}")

    # Write to SQL file
    try:
        with open(output_sql_path, 'w', encoding='utf-8') as f:
            f.write("\n".join(sql_lines))
        print(f"Successfully generated SQL script with {insert_count} rows in {insert_count//BATCH_SIZE + 1} bulk insert statements.")
    except IOError as e:
        print(f"Error writing to SQL file {output_sql_path}: {e}")

if __name__ == "__main__":
    print(f"Generating PostgreSQL bulk insert SQL script '{output_sql_path}' from '{input_json_path}'...")
    print(f"Batch size: {BATCH_SIZE} rows per INSERT statement")
    generate_sql()
    print("Done.")