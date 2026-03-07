import path from "path";
import * as fs from 'fs/promises';
import dotenv from 'dotenv';

import { BandoGISScraper } from "./scrapers/bando-gis.scrapers";

// Load environment variables
dotenv.config();

async function main() {
  console.log("Starting web scraping activity...")

  const gisScraper = new BandoGISScraper();

  try {
    await gisScraper.initialize();

    const result = await gisScraper.scrapeAll();

    // Save results to files
    const outputDir = 'output';
    await fs.mkdir(outputDir, { recursive: true });

    // Save provinces data
    await fs.writeFile(
      path.join(outputDir, 'provinces.json'),
      JSON.stringify(result.provinces, null, 2)
    );

    // Save wards data
    await fs.writeFile(
      path.join(outputDir, 'wards.json'),
      JSON.stringify(result.wards, null, 2)
    );

    // Save complete result
    await fs.writeFile(
      path.join(outputDir, 'complete_result.json'),
      JSON.stringify(result, null, 2)
    );

    // Print summary
    console.log('\n📊 Scraping completed!');
    console.log(`📍 Provinces scraped: ${result.provinces.length}`);
    console.log(`🏘️  Wards scraped: ${result.wards.length}`);
    console.log(`🔬 Total requests: ${result.totalRequests}`);
    console.log(`⏱️  Duration: ${Math.round(result.duration! / 1000)}s`);
    console.log(`💾 Results saved to ./output/ directory`);

    // Failed GIS items summary
    if (result.failedGISItems && result.failedGISItems.length > 0) {
      console.log(`\n❌ Failed GIS Items: ${result.failedGISItems.length}`);
      const failedProvinces = result.failedGISItems.filter(item => item.itemType === 'province');
      const failedWards = result.failedGISItems.filter(item => item.itemType === 'ward');
      
      console.log(`   - Provinces: ${failedProvinces.length}`);
      console.log(`   - Wards: ${failedWards.length}`);
      
      console.log('\n📋 Failed Items Details:');
      result.failedGISItems.forEach((item, index) => {
        const itemName = item.itemData.ten || 'Unknown';
        const itemType = item.itemType === 'province' ? 'Province' : 'Ward';
        console.log(`   ${index + 1}. ${itemType} "${itemName}" - Attempt ${item.attempts} - ${item.lastError || 'Unknown error'}`);
      });
    } else {
      console.log('\n✅ All GIS responses captured successfully!');
    }

    if (result.errors.length > 0) {
      console.log('\n⚠️  Errors encountered:');
      result.errors.forEach((error, index) => {
        console.log(`${index + 1}. ${error}`);
      });
    }

  } catch (error) {
    console.error('💥 Fatal error:', error);
  } finally {
    await gisScraper.cleanup();
  }

}

if (require.main === module) {
  main().catch(console.error);
}