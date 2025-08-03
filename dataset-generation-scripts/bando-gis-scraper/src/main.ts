import path from "path";
import { ScrapingStats } from "./interfaces";
import { BandoGISScraper } from "./scrapers/bando-gis.scrapers";

async function main() {
  console.log("Starting web scraping activity...")

  const gisScraper = new BandoGISScraper();

  const stats: ScrapingStats = {
    totalProvinces: 0,
    totalWards: 0,
    successfulRequests: 0,
    failedRequests: 0,
    startTime: new Date()
  };

  try {
    await gisScraper.initialize();


  } catch (error) {
    console.error('💥 Fatal error:', error);
    stats.endTime = new Date();
    stats.duration = stats.endTime.getTime() - stats.startTime.getTime();
  } finally {
    
  }

}

if (require.main === module) {
  main().catch(console.error);
}
