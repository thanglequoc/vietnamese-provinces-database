import path from "path";
import * as fs from 'fs/promises';

import { BandoGISScraper } from "./scrapers/bando-gis.scrapers";

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

    if (result.errors.length > 0) {
      console.log('\n⚠️  Errors encountered:');
      result.errors.forEach((error, index) => {
        console.log(`${index + 1}. ${error}`);
      });
    }

  } catch (error) {
    console.error('💥 Fatal error:', error);
  } finally {

  }

}

if (require.main === module) {
  main().catch(console.error);
}
