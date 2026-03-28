import json
import os
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
from threading import Lock

# File paths
input_json_path = 'resources/gis/sapnhapbando_15Mar2026/donvi_tinhthanh.json'
output_base_dir = 'resources/gis/sapnhapbando_geojson'

# Configuration
MAX_WORKERS = 10  # Number of parallel requests
API_URL = "https://sapnhap.bando.com.vn/pread_json"

# Lock for thread-safe file operations
file_lock = Lock()

def fetch_and_save_geojson(item):
    """Fetch GeoJSON data from API and save to file"""
    ma = item.get('ma')
    ten = item.get('ten')
    malk = item.get('malk')
    
    if not ma or not ten or not malk:
        return f"Skipping item due to missing fields: {item}"
    
    # Normalize unit name for filename (remove special characters)
    filename = f"{ma}.geojson"
    output_path = os.path.join(output_base_dir, filename)
    
    # Call API
    payload = {"id": malk}
    
    try:
        response = requests.post(API_URL, data=payload, timeout=30)
        
        if response.status_code == 200:
            geojson_data = response.json()
            
            # Save to file (thread-safe)
            with file_lock:
                with open(output_path, 'w', encoding='utf-8') as f_out:
                    json.dump(geojson_data, f_out, ensure_ascii=False, indent=4)
            
            return f"✓ Saved: {ten} ({ma}.geojson)"
        else:
            return f"✗ Failed: {ten} - Status: {response.status_code}"
            
    except requests.exceptions.Timeout:
        return f"✗ Timeout: {ten}"
    except Exception as e:
        return f"✗ Error: {ten} - {str(e)}"

def main():
    """Main function to process all items in parallel"""
    print("Loading JSON data...")
    
    # Load JSON data
    with open(input_json_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Create base output directory if it doesn't exist
    os.makedirs(output_base_dir, exist_ok=True)
    
    total_items = len(data)
    completed = 0
    failed = 0
    
    print(f"Starting parallel download with {MAX_WORKERS} workers...")
    print(f"Total items to fetch: {total_items}\n")
    
    # Use ThreadPoolExecutor for parallel requests
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        # Submit all tasks
        future_to_item = {executor.submit(fetch_and_save_geojson, item): item for item in data}
        
        # Process completed tasks as they complete
        for future in as_completed(future_to_item):
            result = future.result()
            completed += 1
            
            if result.startswith("✗"):
                failed += 1
            
            # Print progress and result
            progress = (completed / total_items) * 100
            print(f"[{completed}/{total_items}] ({progress:.1f}%) {result}")
    
    print(f"\n{'='*60}")
    print(f"Processing complete!")
    print(f"Total: {total_items} | Success: {completed - failed} | Failed: {failed}")
    print(f"{'='*60}")

if __name__ == "__main__":
    main()