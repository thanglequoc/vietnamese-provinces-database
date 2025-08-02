import { Page } from "@playwright/test";
import { ScrapingConfig } from "../interfaces";

export class BrowserService {
  private page: Page | null = null;
  constructor(private config: ScrapingConfig) {}

  async launch(): Promise<void> {
    
  }

}
