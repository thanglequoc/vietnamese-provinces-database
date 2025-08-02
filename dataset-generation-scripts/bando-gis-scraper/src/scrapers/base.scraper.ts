import { Page } from "@playwright/test";
import { APIInterceptorService } from "../services/api-interceptor.service";
import { BrowserService } from "../services/browser.service";
import { SCRAPER_CONFIG } from "../config";

export abstract class BaseScraper {

  protected browserService: BrowserService;
  protected apiInterceptorService: APIInterceptorService;
  protected page: Page | null = null;

  constructor() {
    this.browserService = new BrowserService();
    this.apiInterceptorService = new APIInterceptorService();
  }

  async initialize(): Promise<void> {
    await this.browserService.initialize();
    this.page = this.browserService.getPage();
    await this.apiInterceptorService.setupInterceptor(this.page);
    await this.navigateToSapNhapBanDoSite();
  }

  protected async navigateToSapNhapBanDoSite(): Promise<void> {
    if (!this.page) throw new Error('Page is not initialized');

    await this.page.goto(SCRAPER_CONFIG.BASE_URL, {
      waitUntil: 'networkidle',
      timeout: SCRAPER_CONFIG.TIMEOUTS.PAGE_LOAD
    });

    await this.page.waitForSelector(SCRAPER_CONFIG.SELECTORS.PROVINCE_TABLE, {
      timeout: SCRAPER_CONFIG.TIMEOUTS.ELEMENT_WAIT
    })
  }

  protected async waitForNetworkIdle(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForLoadState('networkidle')
    await this.page.waitForTimeout(SCRAPER_CONFIG.TIMEOUTS.REQUEST_WAIT)
  }

  protected async retryOperation<T>(
    operation: () => Promise<T>,
    maxAttempts: number = SCRAPER_CONFIG.RETRY.MAX_ATTEMPTS
  ): Promise<T> {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;
        console.warn(`Attempt ${attempt} failed:`, error);

        if (attempt < maxAttempts) {
          await this.page?.waitForTimeout(SCRAPER_CONFIG.RETRY.DELAY);
        }
      }
    }

    throw lastError!;
  }

  async cleanup(): Promise<void> {
    await this.browserService.close();
  }
}
