import path from "path";
import { DEFAULT_CONFIG } from "./config";
import { ScrapingConfig } from "./interfaces";

async function main() {
  console.log("Starting web scraping activity...")

  const config: ScrapingConfig = {
    ...DEFAULT_CONFIG,
    websiteUrl: process.env.TARGET_WEBSITE_URL || 'YOUR_WEBSITE_URL_HERE',
    headless: process.env.NODE_ENV === 'production',
    outputFile: path.join('output', 'vietnamese_gis_complete.json')
  };

  
}

if (require.main === module) {
  main().catch(console.error);
}
