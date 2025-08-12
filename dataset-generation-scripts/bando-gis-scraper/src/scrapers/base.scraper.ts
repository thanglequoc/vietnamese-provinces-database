import { Locator, Page } from "@playwright/test";
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

  /**
   * Scrolls through a Tabulator virtual table to collect all items
   * Handles padding-top/padding-bottom changes in virtual scrolling
   */
  protected async scrollAndCollectAllItems<T>(
    tableSelector: string,
    rowSelector: string,
    itemExtractor: (row: Locator) => Promise<T>,
    containerSelector?: string
  ): Promise<T[]> {
    if (!this.page) throw new Error('Page not initialized');

    const allItems: T[] = [];
    const seenItems = new Set<string>();
    let scrollAttempts = 0;
    let stableScrollAttempts = 0;
    let lastPaddingTop = -1;
    let lastPaddingBottom = -1;

    console.log(`📜 Starting Tabulator virtual scroll collection from ${tableSelector}`);

    // Find the scrollable container
    const scrollContainer = containerSelector 
      ? this.page.locator(containerSelector).first()
      : this.page.locator(SCRAPER_CONFIG.SELECTORS.TABLE_HOLDER).first();

    // Reset to top
    await this.scrollToTop(scrollContainer);
    await this.page.waitForTimeout(SCRAPER_CONFIG.TIMEOUTS.STABILIZATION_WAIT);

    while (scrollAttempts < SCRAPER_CONFIG.SCROLLING.MAX_SCROLL_ATTEMPTS) {
      const table = this.page.locator(tableSelector).first();
      
      // Get current padding values to understand scroll position
      const paddingInfo = await table.evaluate((el) => {
        const style = window.getComputedStyle(el);
        return {
          paddingTop: parseInt(style.paddingTop) || 0,
          paddingBottom: parseInt(style.paddingBottom) || 0,
          totalHeight: parseInt(style.paddingTop) + parseInt(style.paddingBottom) + el.scrollHeight
        };
      });

      console.log(`📊 Scroll ${scrollAttempts + 1}: padding-top=${paddingInfo.paddingTop}px, padding-bottom=${paddingInfo.paddingBottom}px`);

      // Check if we've reached the end (padding-bottom becomes 0 and padding-top is at max)
      if (paddingInfo.paddingBottom === 0 && scrollAttempts > 0) {
        console.log('🏁 Reached end of table (padding-bottom = 0)');
        stableScrollAttempts++;
        if (stableScrollAttempts >= 2) break;
      }

      // Extract items from currently visible rows
      const rows = await table.locator(rowSelector).all();
      let newItemsFound = 0;

      for (const row of rows) {
        try {
          const item = await itemExtractor(row);
          const itemKey = this.generateItemKey(item);
          
          if (!seenItems.has(itemKey)) {
            seenItems.add(itemKey);
            allItems.push(item);
            newItemsFound++;
          }
        } catch (error) {
          console.warn('Error extracting item from row:', error);
        }
      }

      console.log(`✅ Found ${newItemsFound} new items (Total: ${allItems.length}) from ${rows.length} visible rows`);

      // Check if padding values haven't changed (indicating we might be stuck)
      if (paddingInfo.paddingTop === lastPaddingTop && paddingInfo.paddingBottom === lastPaddingBottom) {
        stableScrollAttempts++;
        if (stableScrollAttempts >= SCRAPER_CONFIG.SCROLLING.DUPLICATE_CHECK_THRESHOLD) {
          console.log('🔄 Padding values stable, assuming end reached');
          break;
        }
      } else {
        stableScrollAttempts = 0;
      }

      lastPaddingTop = paddingInfo.paddingTop;
      lastPaddingBottom = paddingInfo.paddingBottom;

      // Scroll down using Tabulator-specific methods
      await this.scrollTabulatorDown(scrollContainer, table);
      await this.page.waitForTimeout(SCRAPER_CONFIG.SCROLLING.SCROLL_DELAY);

      // Wait for virtual scrolling to update
      await this.waitForTabulatorUpdate(table, paddingInfo.paddingTop);

      scrollAttempts++;
    }

    // Reset to top for subsequent operations
    await this.scrollToTop(scrollContainer);
    await this.page.waitForTimeout(SCRAPER_CONFIG.TIMEOUTS.STABILIZATION_WAIT);

    console.log(`📋 Collected ${allItems.length} unique items after ${scrollAttempts} scroll attempts`);
    return allItems;
  }

  /**
   * Scrolls down in a Tabulator virtual table using multiple strategies
   */
  private async scrollTabulatorDown(container: Locator, table: Locator): Promise<void> {
    try {
      // Method 1: Scroll the container
      await container.evaluate((el, step) => {
        el.scrollTop += step;
      }, SCRAPER_CONFIG.SCROLLING.SCROLL_STEP);

      // Method 2: Use Tabulator's internal scrolling if available
      await table.evaluate((el, step) => {
        const tabulatorEl = el.closest('.tabulator');
        if (tabulatorEl && (tabulatorEl as any).tabulator) {
          // Use Tabulator's scrollToRow if we can estimate row
          const currentPaddingTop = parseInt(getComputedStyle(el).paddingTop) || 0;
          const estimatedRowHeight = 29; // Based on your HTML
          const currentRow = Math.floor(currentPaddingTop / estimatedRowHeight);
          const nextRows = Math.ceil(step / estimatedRowHeight);
          
          try {
            (tabulatorEl as any).tabulator.scrollToRow(currentRow + nextRows, "top", false);
          } catch (e) {
            // Fallback to container scroll
            const container = el.closest('.tabulator-tableholder');
            if (container) container.scrollTop += step;
          }
        }
      }, SCRAPER_CONFIG.SCROLLING.SCROLL_STEP);

      // Method 3: Keyboard navigation fallback
      await container.press('PageDown');

    } catch (error) {
      console.warn('Error in Tabulator scrolling:', error);
      // Final fallback: mouse wheel
      await this.page?.mouse.wheel(0, SCRAPER_CONFIG.SCROLLING.SCROLL_STEP);
    }
  }

  /**
   * Waits for Tabulator virtual scrolling to update the padding values
   */
  private async waitForTabulatorUpdate(table: Locator, previousPaddingTop: number): Promise<void> {
    let attempts = 0;
    const maxAttempts = 10;

    while (attempts < maxAttempts) {
      await this.page?.waitForTimeout(SCRAPER_CONFIG.SCROLLING.VIRTUAL_SCROLL_WAIT);

      try {
        const currentPaddingTop = await table.evaluate((el) => {
          return parseInt(window.getComputedStyle(el).paddingTop) || 0;
        });

        // If padding changed or we've waited enough, consider it updated
        if (currentPaddingTop !== previousPaddingTop || attempts >= 3) {
          break;
        }
      } catch (error) {
        break;
      }

      attempts++;
    }
  }

  /**
   * Enhanced scroll to top for Tabulator
   */
  private async scrollToTop(container: Locator): Promise<void> {
    try {
      // Method 1: Direct container scroll
      await container.evaluate(el => {
        el.scrollTop = 0;
      });

      // Method 2: Use Tabulator API if available
      await container.evaluate(el => {
        const tabulatorEl = el.closest('.tabulator');
        if (tabulatorEl && (tabulatorEl as any).tabulator) {
          try {
            (tabulatorEl as any).tabulator.scrollToRow(1, "top", false);
          } catch (e) {
            // Ignore if scrollToRow fails
          }
        }
      });

      // Method 3: Keyboard shortcut
      await container.press('Home');

    } catch (error) {
      console.warn('Error scrolling to top:', error);
    }
  }

  /**
   * Generates a unique key for an item to detect duplicates
   */
  private generateItemKey(item: any): string {
    if (typeof item === 'object' && item !== null) {
      // Create key from main identifying properties
      const { stt, ten, truocsn } = item;
      return `${stt || ''}_${ten || ''}_${(truocsn || '').substring(0, 50)}`;
    }
    return JSON.stringify(item);
  }

  /**
   * Finds a specific item in the currently visible rows by matching properties
   */
  protected async findVisibleItemIndex<T>(
    tableSelector: string,
    rowSelector: string,
    itemExtractor: (row: Locator) => Promise<T>,
    targetItem: T
  ): Promise<number> {
    if (!this.page) throw new Error('Page not initialized');

    const table = this.page.locator(tableSelector).first();
    const rows = await table.locator(rowSelector).all();
    
    const targetKey = this.generateItemKey(targetItem);

    for (let i = 0; i < rows.length; i++) {
      try {
        const currentItem = await itemExtractor(rows[i]);
        const currentKey = this.generateItemKey(currentItem);
        
        if (currentKey === targetKey) {
          return i;
        }
      } catch (error) {
        console.warn(`Error checking row ${i}:`, error);
      }
    }

    return -1; // Not found in current view
  }

  /**
   * Scrolls to find a specific item in Tabulator virtual table and returns its current index
   */
  protected async scrollToItem<T>(
    tableSelector: string,
    rowSelector: string,
    itemExtractor: (row: Locator) => Promise<T>,
    targetItem: T,
    containerSelector?: string
  ): Promise<number> {
    if (!this.page) throw new Error('Page not initialized');

    const scrollContainer = containerSelector 
      ? this.page.locator(containerSelector).first()
      : this.page.locator(SCRAPER_CONFIG.SELECTORS.TABLE_HOLDER).first();

    const table = this.page.locator(tableSelector).first();
    const targetKey = this.generateItemKey(targetItem);

    // First, try to find in current view
    let itemIndex = await this.findVisibleItemIndex(tableSelector, rowSelector, itemExtractor, targetItem);
    if (itemIndex !== -1) {
      return itemIndex;
    }

    // Reset to top and search
    await this.scrollToTop(scrollContainer);
    await this.page.waitForTimeout(SCRAPER_CONFIG.TIMEOUTS.STABILIZATION_WAIT);

    let scrollAttempts = 0;
    let lastPaddingTop = -1;

    while (scrollAttempts < SCRAPER_CONFIG.SCROLLING.MAX_SCROLL_ATTEMPTS) {
      // Check current visible items
      itemIndex = await this.findVisibleItemIndex(tableSelector, rowSelector, itemExtractor, targetItem);
      if (itemIndex !== -1) {
        console.log(`🎯 Found target item at visible index ${itemIndex} after ${scrollAttempts} scrolls`);
        return itemIndex;
      }

      // Get current padding to understand position
      const paddingInfo = await table.evaluate((el) => ({
        paddingTop: parseInt(window.getComputedStyle(el).paddingTop) || 0,
        paddingBottom: parseInt(window.getComputedStyle(el).paddingBottom) || 0
      }));

      // Stop if we've reached the end
      if (paddingInfo.paddingBottom === 0 && scrollAttempts > 0) {
        break;
      }

      // Stop if padding isn't changing (stuck)
      if (paddingInfo.paddingTop === lastPaddingTop && scrollAttempts > 3) {
        break;
      }

      lastPaddingTop = paddingInfo.paddingTop;

      // Scroll down
      await this.scrollTabulatorDown(scrollContainer, table);
      await this.page.waitForTimeout(SCRAPER_CONFIG.SCROLLING.SCROLL_DELAY);
      await this.waitForTabulatorUpdate(table, paddingInfo.paddingTop);

      scrollAttempts++;
    }

    throw new Error(`Could not find target item "${targetKey}" after ${scrollAttempts} scroll attempts`);
  }

  async cleanup(): Promise<void> {
    await this.browserService.close();
  }
}
