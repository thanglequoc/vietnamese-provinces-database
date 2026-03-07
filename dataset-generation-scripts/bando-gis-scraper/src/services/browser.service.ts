import { Browser, BrowserContext, chromium, Page } from "@playwright/test";

export class BrowserService {
  private browser: Browser | null = null;
  private context: BrowserContext | null = null;
  private page: Page | null = null;

  async initialize(): Promise<void> {
    // Read HEADLESS from environment, default to false for backward compatibility
    const headless = process.env.HEADLESS === 'true';
    
    this.browser = await chromium.launch({
      headless,
      slowMo: 200,
    })

    this.context = await this.browser.newContext({
      viewport: { width: 1920, height: 1080 },
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
    });

    this.page = await this.context.newPage();
  }

  getPage(): Page {
    if (!this.page) {
      throw new Error('Browser not initialized. Call initialize() first.');
    }
    return this.page;
  }

  async close(): Promise<void> {
    if (this.page) await this.page.close();
    if (this.context) await this.context.close();
    if (this.browser) await this.browser.close();
  }

}
