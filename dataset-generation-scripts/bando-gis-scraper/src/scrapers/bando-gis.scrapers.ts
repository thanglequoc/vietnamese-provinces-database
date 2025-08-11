import { BaseScraper } from "./base.scraper";

import { ProvinceData } from "../interfaces/scraper.interfaces"
import { SCRAPER_CONFIG } from "../config";

// TODO @thangle: Implement the GIS scraper
export class BandoGISScraper extends BaseScraper {
  async scrapeAll(): Promise<void> {
    // TODO @thangle: Implement the scrape all function

    const provinces = await this.getProvinceList()
    console.log(`Found ${provinces.length} provinces`);

  }

  private async getProvinceList(): Promise<ProvinceData[]> {
    if (!this.page) throw new Error('Page not initialized');

    const provinces: ProvinceData[] = [];

    const provinceTableSelector = await this.page.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE).all();
    if (provinceTableSelector.length === 0) {
      throw new Error('Unable to locate the province table');
    }
    const provinceTable = provinceTableSelector[0];

    const rows = await provinceTable.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_ROW).all();
    for (const row of rows) {
      const cells = await row.locator(SCRAPER_CONFIG.SELECTORS.CELL).all();

      if (cells.length >= 3) {
        const stt = await cells[0].textContent() || '';
        const ten = await cells[1].textContent() || '';
        const truocsn = await cells[2].textContent() || '';

        provinces.push({ stt: stt.trim(), ten: ten.trim(), truocsn: truocsn.trim() });
      }
    }

    return provinces;
  }
}