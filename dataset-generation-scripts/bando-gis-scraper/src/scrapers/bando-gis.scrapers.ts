import { BaseScraper } from "./base.scraper";

import { ProvinceData, WardData } from "../interfaces/scraper.interfaces"
import { SCRAPER_CONFIG } from "../config";
import { Locator } from "@playwright/test";
import { ProvinceGISServerResponse, ResponseType } from "../interfaces";

// TODO @thangle: Implement the GIS scraper
export class BandoGISScraper extends BaseScraper {
  async scrapeAll(): Promise<void> {
    // TODO @thangle: Implement the scrape all function

    const provinces = await this.getProvinceList()
    console.log(`Found ${provinces.length} provinces`);

    for (let i = 0; i < provinces.length; i++) {
      const province = provinces[i];
      console.log(`Processing province ${i + 1}/${provinces.length}: ${province.ten}`);

      try {
        // click on the province and get both GIS and info data
        const provinceData = await this.clickProvinceAndGetGIS(i, provinces[i]);


        // Get wards for this province
        const wards = await this.getWardList();
        console.log(`Found ${wards.length} wards for ${province.ten}`);
      } catch (err) {
        console.log("TODO @thangle: Improving...")
      }
    }
  }

  private async getWardList(): Promise<WardData[]> {
    if (!this.page) throw new Error('Page not initialized');
    console.log('🏘️ Collecting all wards from Tabulator virtual table...');

    const itemExtractor = async (row: Locator): Promise<WardData> => {
      const cells = await row.locator(SCRAPER_CONFIG.SELECTORS.CELL).all();
      if (cells.length >= 3) {
        const stt = await cells[0].textContent() || '';
        const ten = await cells[1].textContent() || '';
        const truocsn = await cells[2].textContent() || '';

        return { stt: stt.trim(), ten: ten.trim(), truocsn: truocsn.trim() };
      }
      throw new Error('Invalid row structure');
    };

    const wardTables = await this.page.locator(SCRAPER_CONFIG.SELECTORS.WARD_TABLE).all();
    if (wardTables.length === 0) {
      throw new Error('Ward table not found');
    }

    const wards = await this.scrollAndCollectAllItems(
      SCRAPER_CONFIG.SELECTORS.WARD_TABLE,
      SCRAPER_CONFIG.SELECTORS.WARD_ROW,
      itemExtractor,
      SCRAPER_CONFIG.SELECTORS.WARD_TABLE_HOLDER
    );

    return wards;
  }

  private async clickProvinceAndGetGIS(provinceIndex: number, targetProvince: ProvinceData): Promise<{ gisData?: ProvinceGISServerResponse }> {
    if (!this.page) throw new Error('Page not initialized');

    return await this.retryOperation(async () => {
      // Set user action context and clear previous interceptor tracking data
      this.apiInterceptorService.setUserAction('province_click');
      this.apiInterceptorService.clearInterceptedData();

      const timestamp = Date.now();

      // Scroll to make the target province visible and get its current index
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

      const visibleIndex = await this.scrollToItem(
        SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE,
        SCRAPER_CONFIG.SELECTORS.PROVINCE_ROW,
        itemExtractor,
        targetProvince,
        SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE_HOLDER
      );

      // Now click on the visible row
      const provinceTables = await this.page!.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE).all();
      const provinceTable = provinceTables[0];
      const rows = await provinceTable.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_ROW).all();

      if (visibleIndex >= rows.length) {
        throw new Error(`Province visible index ${visibleIndex} out of bounds`);
      }

      console.log(`🎯 Clicking on province "${targetProvince.ten}" at visible index ${visibleIndex}`);
      await rows[visibleIndex].click();
      await this.waitForNetworkIdle();

      // Get responsed that came after the click
      const gisResponses = this.apiInterceptorService.getResponsesSince(timestamp, ResponseType.PROVINCE_GIS);

      return {
        gisData: gisResponses.length > 0 ? gisResponses[gisResponses.length - 1].response : undefined,
      }
    })
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
      SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE_HOLDER
    );

    return provinces;
  }
}