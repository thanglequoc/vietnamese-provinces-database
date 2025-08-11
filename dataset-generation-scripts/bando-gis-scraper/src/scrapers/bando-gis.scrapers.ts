import { BaseScraper } from "./base.scraper";

import { ProvinceData } from "../interfaces/scraper.interfaces"
import { SCRAPER_CONFIG } from "../config";
import { Locator } from "@playwright/test";

// TODO @thangle: Implement the GIS scraper
export class BandoGISScraper extends BaseScraper {
  async scrapeAll(): Promise<void> {
    // TODO @thangle: Implement the scrape all function

    const provinces = await this.getProvinceList()
    console.log(`Found ${provinces.length} provinces`);

    

  }

  private async getProvinceList(): Promise<ProvinceData[]> {
    if (!this.page) throw new Error('Page not initialized');

    console.log('🏛️ Collecting all provinces from Tabulator virtual table...');

    const itemExtractor = async (row: Locator): Promise<ProvinceData> => {
      const cells = await row.locator(SCRAPER_CONFIG.SELECTORS.CELL).all();
      if (cells.length >= 3) {
        const stt = await cells[0].textContent() || '';
        const ten = await cells[1].textContent() || '';
        const truocsn = await cells[2].textContent() || '';

        return { stt: stt.trim(), ten: ten.trim(), truocsn: truocsn.trim() };
      }
      throw new Error('Invalid row structure');
    };

    // Get the first table (provinces) and scroll through it
    const provinceTables = await this.page.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE).all();
    if (provinceTables.length === 0) {
      throw new Error('Province table not found');
    }

    // Use the Tabulator-aware scrolling method from base class
    const provinces = await this.scrollAndCollectAllItems(
      SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE, // Target first table
      SCRAPER_CONFIG.SELECTORS.PROVINCE_ROW,
      itemExtractor,
      SCRAPER_CONFIG.SELECTORS.TABLE_HOLDER
    );

    return provinces;
  }
}