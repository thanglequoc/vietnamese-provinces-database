# SAPNHAP Bando Crawler

This directory contains Python scripts for crawling and processing data from the Sáp Nhập Bando (Vietnamese administrative unit consolidation) system. The scripts are designed to fetch GeoJSON data and generate SQL database tables from the Bando API.

## Overview

The Sáp Nhập Bando (Administrative Unit Consolidation) system provides geographic and administrative data for Vietnamese provinces, districts, and wards. These scripts help you:

1. **Fetch GeoJSON data** in parallel from the Bando API
2. **Generate PostgreSQL SQL scripts** from JSON data with bulk insert optimizations

## Files

### Scripts

- **`generate_geojson_from_sapnhap.py`** - Fetches GeoJSON data from the Bando API in parallel and saves individual GeoJSON files
- **`generate_sql_sapnhap.py`** - Converts JSON data into PostgreSQL bulk insert SQL statements

### Data Files

- **`donvi_tinhthanh.json`** - Input JSON file containing administrative unit data (ma, ten, magoc, malk, truocsapnhap)
- **`bando_co_dvch.sql`** - Generated SQL output file containing INSERT statements for PostgreSQL

## Prerequisites

### Python Requirements

- Python 3.6 or higher
- Required Python packages:
  ```bash
  pip install requests
  ```

### Input Data Format

The scripts expect a JSON file with the following structure:

```json
[
  {
    "ma": "01",
    "ten": "Thành phố Hà Nội",
    "magoc": "123",
    "malk": "456",
    "truocsapnhap": "Thủ đô Hà Nội"
  },
  ...
]
```

**Field Descriptions:**
- `ma`: Administrative unit code (required)
- `ten`: Administrative unit name (required)
- `magoc`: Original/parent code (optional, set to "0" for NULL)
- `malk`: API identifier used to fetch GeoJSON (required for GeoJSON generation)
- `truocsapnhap`: Name before consolidation (optional)

## Usage

### 1. Generate SQL from JSON

Convert the JSON data into PostgreSQL bulk insert statements:

```bash
python generate_sql_sapnhap.py
```

**What it does:**
- Reads from `./donvi_tinhthanh.json`
- Generates `bando_co_dvch.sql` with bulk INSERT statements
- Uses batch size of 1000 rows per INSERT for optimal performance
- Cleans strings by removing quotes and extra spaces
- Handles `magoc` values of "0" as NULL

**Output:**
The generated SQL file (`bando_co_dvch.sql`) can be imported into PostgreSQL:
```bash
psql -U your_username -d your_database -f bando_co_dvch.sql
```

### 2. Fetch GeoJSON Data (Optional)

Fetch GeoJSON data from the Bando API in parallel:

```bash
python generate_geojson_from_sapnhap.py
```

**What it does:**
- Reads from `resources/gis/sapnhapbando_15Mar2026/donvi_tinhthanh.json`
- Makes parallel API requests to `https://sapnhap.bando.com.vn/pread_json`
- Saves individual GeoJSON files to `resources/gis/sapnhapbando_geojson/`
- Uses 10 concurrent workers for efficient fetching
- Provides real-time progress updates

**Configuration:**
- `MAX_WORKERS = 10` - Number of parallel requests (adjust based on your network and API limits)
- `API_URL = "https://sapnhap.bando.com.vn/pread_json"` - API endpoint

**Output:**
GeoJSON files named after unit codes (e.g., `01.geojson`, `79.geojson`) in the output directory.

## Configuration

### `generate_sql_sapnhap.py`

Edit the following variables in the script:

```python
# File paths
input_json_path = './donvi_tinhthanh.json'  # Input JSON file
output_sql_path = 'bando_co_dvch.sql'        # Output SQL file

# Configuration
BATCH_SIZE = 1000  # Rows per INSERT statement
```

### `generate_geojson_from_sapnhap.py`

Edit the following variables in the script:

```python
# File paths
input_json_path = 'resources/gis/sapnhapbando_15Mar2026/donvi_tinhthanh.json'
output_base_dir = 'resources/gis/sapnhapbando_geojson'

# Configuration
MAX_WORKERS = 10  # Number of parallel requests
API_URL = "https://sapnhap.bando.com.vn/pread_json"
```

## Database Table Schema

Before importing the generated SQL, you need to create the table. Use the `create-table-sapnhap-bando-data.sql` script which creates the following schema:

```sql
CREATE TABLE bando_co_dvch (
    ma TEXT PRIMARY KEY,
    ten TEXT NOT NULL,
    magoc TEXT,
    malk TEXT,
    truocsapnhap TEXT,
    
    CONSTRAINT fk_bando_co_dvch_parent
        FOREIGN KEY (magoc)
        REFERENCES bando_co_dvch(ma)
        ON DELETE SET NULL
        ON UPDATE CASCADE
);
```

**Column Descriptions:**
- `ma` (TEXT, PRIMARY KEY) - Administrative unit code (e.g., 'ti33', 'ti21')
- `ten` (TEXT, NOT NULL) - Administrative unit name (e.g., 'Thành Phố Cần Thơ')
- `magoc` (TEXT, nullable) - Original/parent code (self-referencing foreign key to `ma`)
- `malk` (TEXT, nullable) - API identifier used to fetch GeoJSON (e.g., 'diaphanhanhchinhcaptinh_sn.1')
- `truocsapnhap` (TEXT, nullable) - Name before consolidation (e.g., 'thành phố Cần Thơ, tỉnh Sóc Trăng và tỉnh Hậu Giang')

## Error Handling

### SQL Generation
- Validates required fields (`ma`, `ten`)
- Skips items with missing required fields
- Handles JSON decode errors gracefully
- Reports errors if file operations fail

### GeoJSON Fetching
- Handles network timeouts (30 second timeout per request)
- Catches and reports exceptions for each request
- Provides detailed success/failure status for each item
- Tracks and reports total success/failure counts

## Performance Considerations

### SQL Generation
- Uses bulk INSERT statements (batch size: 1000 rows)
- Significantly faster than individual INSERT statements
- Reduces database round-trips

### GeoJSON Fetching
- Parallel processing with 10 concurrent workers
- Thread-safe file operations using locks
- Progress tracking with real-time updates

**Tip:** Adjust `MAX_WORKERS` based on:
- Network bandwidth
- API rate limits
- System resources

## Example Workflow

1. **Prepare input data:**
   Ensure `donvi_tinhthanh.json` contains the administrative unit data with all required fields.

2. **Generate SQL:**
   ```bash
   python generate_sql_sapnhap.py
   ```

3. **Create table in database:**
   ```bash
   psql -U postgres -d vietnam_provinces -f create-table-sapnhap-bando-data.sql
   ```

4. **Import data to database:**
   ```bash
   psql -U postgres -d vietnam_provinces -f bando_co_dvch.sql
   ```

5. **(Optional) Fetch GeoJSON:**
   ```bash
   python generate_geojson_from_sapnhap.py
   ```

**Note:** Steps 3 and 4 assume you have PostgreSQL installed and configured. Replace `postgres` and `vietnam_provinces` with your actual database username and database name.

## Troubleshooting

### Issue: "Input file not found"
**Solution:** Ensure the input JSON file exists at the specified path. Check the `input_json_path` variable in the script.

### Issue: "Invalid JSON in file"
**Solution:** Validate your JSON file using a JSON validator or `python -m json.tool donvi_tinhthanh.json`.

### Issue: Timeout errors when fetching GeoJSON
**Solution:** Reduce `MAX_WORKERS` to decrease concurrent requests, or increase the `timeout` parameter in the `requests.post()` call.

### Issue: Database import errors
**Solution:** Ensure the target table exists or modify the SQL script to include `CREATE TABLE` statements if needed.

## Notes

- The scripts use UTF-8 encoding for reading/writing files
- SQL generation handles string escaping properly (single quotes → double quotes)
- The GeoJSON fetching script creates the output directory automatically if it doesn't exist
- Both scripts provide console output for progress tracking

## License

These scripts are part of the Vietnamese Provinces Database project. Please refer to the main repository for licensing information.