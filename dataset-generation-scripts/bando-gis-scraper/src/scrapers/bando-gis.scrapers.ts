import { BaseScraper } from "./base.scraper";

import { ProvinceData, WardData, FailedGISItem } from "../interfaces/scraper.interfaces"
import { SCRAPER_CONFIG } from "../config";
import { Locator } from "@playwright/test";
import { APIInterceptedRequest, ResponseType, ScrapingResult } from "../interfaces";

export class BandoGISScraper extends BaseScraper {
  async scrapeAll(): Promise<ScrapingResult> {
    const result: ScrapingResult = {
      provinces: [],
      wards: [],
      totalRequests: 0,
      errors: [],
      startTime: new Date(),
      failedGISItems: [],
    };

    const provinces = await this.getProvinceList()
    console.log(`Found ${provinces.length} provinces`);
    this.apiInterceptorService.clearInterceptedData(ResponseType.PROVINCE_GIS);
    this.apiInterceptorService.clearInterceptedData(ResponseType.WARD_GIS);

    for (let i = 0; i < provinces.length; i++) {
      const province = provinces[i];
      console.log(`Processing province ${i + 1}/${provinces.length}: ${province.ten}`);

      try {
        // click on province and get both GIS and info data
        const provinceGISData = await this.clickProvinceAndGetGIS(i, provinces[i]);
        console.log(`Province ${province.ten} - GIS: ${JSON.stringify(provinceGISData)}`)
        province.gisServerResponse = JSON.stringify(provinceGISData);
        result.provinces.push(province);

        // Get wards for this province
        const wards = await this.getWardList();
        console.log(`Found ${wards.length} wards for ${province.ten}`);
        result.totalRequests++;

        // Scrape each ward
        for (let j = 0; j < wards.length; j++) {
          const ward = wards[j]
          ward.provinceName = province.ten

          try {
            const wardGISData = await this.clickWardAndGetGIS(j, ward)
            ward.gisServerResponse = JSON.stringify(wardGISData);
            result.wards.push(ward);
            result.totalRequests++;
            console.log(`Ward ${ward.ten} - GIS: ${JSON.stringify(wardGISData)}`)
          } catch (error) {
            // Track failed ward GIS capture
            const failedItem: FailedGISItem = {
              itemType: 'ward',
              itemData: ward,
              attempts: SCRAPER_CONFIG.RETRY.GIS_MAX_ATTEMPTS,
              lastError: error instanceof Error ? error.message : String(error),
              timestamp: new Date()
            };
            result.failedGISItems.push(failedItem);
            
            const errorMsg = `Error scraping ward ${ward.ten}: ${error}`;
            console.error(errorMsg);
            result.errors.push(errorMsg);
          }
        }
      } catch (error) {
        // Track failed province GIS capture
        const failedItem: FailedGISItem = {
          itemType: 'province',
          itemData: province,
          attempts: SCRAPER_CONFIG.RETRY.GIS_MAX_ATTEMPTS,
          lastError: error instanceof Error ? error.message : String(error),
          timestamp: new Date()
        };
        result.failedGISItems.push(failedItem);
        
        const errorMsg = `Error scraping province ${province.ten}: ${error}`;
        console.error(errorMsg);
        result.errors.push(errorMsg);
      }

      await this.page?.waitForTimeout(SCRAPER_CONFIG.TIMEOUTS.BETWEEN_CLICKS);
    }

    result.endTime = new Date();
    result.duration = result.endTime.getTime() - result.startTime.getTime();
    return result;
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

  private async clickProvinceAndGetGIS(provinceIndex: number, targetProvince: ProvinceData): Promise<APIInterceptedRequest> {
    if (!this.page) throw new Error('Page not initialized');

    return await this.retryOperationWithValidation(
      async () => {
        // Set user action context and clear previous interceptor tracking data
        this.apiInterceptorService.setUserAction('province_click');
        this.apiInterceptorService.clearInterceptedData();

        const timestamp = Date.now();

        // Scroll to make target province visible and get its current index
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

        // Now click on visible row
        const provinceTables = await this.page!.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE).all();
        const provinceTable = provinceTables[0];
        const rows = await provinceTable.locator(SCRAPER_CONFIG.SELECTORS.PROVINCE_ROW).all();

        if (visibleIndex >= rows.length) {
          throw new Error(`Province visible index ${visibleIndex} out of bounds`);
        }

        console.log(`🎯 Clicking on province "${targetProvince.ten}" at visible index ${visibleIndex}`);
        await rows[visibleIndex].click();
        await this.waitForNetworkIdle();

        // Get responses that came after the click
        const gisResponses = this.apiInterceptorService.getResponsesSince(timestamp, ResponseType.PROVINCE_GIS);

        if (gisResponses.length === 0) {
          throw new Error('No GIS response captured for province');
        }

        return gisResponses[gisResponses.length - 1].response;
      },
      // Validator: Check if response exists and is not null/undefined
      (result) => result !== null && result !== undefined,
      SCRAPER_CONFIG.RETRY.GIS_MAX_ATTEMPTS,
      `Province "${targetProvince.ten}"`
    );
  }

  private async clickWardAndGetGIS(wardIndex: number, targetWard: WardData): Promise<APIInterceptedRequest> {
    if (!this.page) throw new Error('Page not initialized');

    return await this.retryOperationWithValidation(
      async () => {
        // Set user action context and clear previous data
        this.apiInterceptorService.setUserAction('ward_click');
        this.apiInterceptorService.clearInterceptedData();

        const timestamp = Date.now();

        // Scroll to make target ward visible and get its current index
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

        const visibleIndex = await this.scrollToItem(
          SCRAPER_CONFIG.SELECTORS.WARD_TABLE,
          SCRAPER_CONFIG.SELECTORS.WARD_ROW,
          itemExtractor,
          targetWard,
          SCRAPER_CONFIG.SELECTORS.WARD_TABLE_HOLDER
        );

        // Now click on visible row
        const wardTables = await this.page!.locator(SCRAPER_CONFIG.SELECTORS.WARD_TABLE).all();
        const wardTable = wardTables[0];
        const rows = await wardTable.locator(SCRAPER_CONFIG.SELECTORS.WARD_ROW).all();

        if (visibleIndex >= rows.length) {
          throw new Error(`Ward visible index ${visibleIndex} out of bounds`);
        }

        console.log(`🎯 Clicking on ward "${targetWard.ten}" at visible index ${visibleIndex}`);
        await rows[visibleIndex].click();
        await this.waitForNetworkIdle();

        // Get responses that came after the click
        const gisResponses = this.apiInterceptorService.getResponsesSince(timestamp, ResponseType.WARD_GIS);

        if (gisResponses.length === 0) {
          throw new Error('No GIS response captured for ward');
        }

        return gisResponses[gisResponses.length - 1].response;
      },
      // Validator: Check if response exists and is not null/undefined
      (result) => result !== null && result !== undefined,
      SCRAPER_CONFIG.RETRY.GIS_MAX_ATTEMPTS,
      `Ward "${targetWard.ten}"`
    );
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

    // Get first table (provinces) and scroll through it
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