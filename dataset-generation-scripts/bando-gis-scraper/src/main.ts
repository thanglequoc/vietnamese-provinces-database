import path from "path";
import { BandoGISScraper } from "./scrapers/bando-gis.scrapers";

async function main() {
  console.log("Starting web scraping activity...")

  const gisScraper = new BandoGISScraper();

  try {
    await gisScraper.initialize();

    const result = gisScraper.scrapeAll();
  } catch (error) {
    console.error('💥 Fatal error:', error);
  } finally {
    
  }

}

if (require.main === module) {
  main().catch(console.error);
}
