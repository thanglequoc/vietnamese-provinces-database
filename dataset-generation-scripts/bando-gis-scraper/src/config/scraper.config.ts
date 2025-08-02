import { ScrapingConfig } from "../interfaces/scraper.intefaces";

export const DEFAULT_CONFIG: Required<ScrapingConfig> = {
  websiteUrl: process.env.BANDO_WEBSITE_URL || '',
  headless: process.env.NODE_ENV === 'production',
  delayBetweenClicks: parseInt(process.env.DELAY_BETWEEN_CLICKS || '1500'),
  delayBetweenProvinces: parseInt(process.env.DELAY_BETWEEN_PROVINCES || '3000'),
  outputFile: process.env.OUTPUT_FILE || 'vietnamese_gis_data.json',
  maxRetries: parseInt(process.env.MAX_RETRIES || '3'),
  timeout: parseInt(process.env.TIMEOUT || '30000'),
  userAgent: process.env.USER_AGENT || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
  viewport: {
    width: parseInt(process.env.VIEWPORT_WIDTH || '1400'),
    height: parseInt(process.env.VIEWPORT_HEIGHT || '900')
  }
};

export const BROWSER_OPTIONS = {
  slowMo: 50,
  args: [
    '--no-sandbox',
    '--disable-setuid-sandbox',
    '--disable-dev-shm-usage',
    '--disable-accelerated-2d-canvas',
    '--no-first-run',
    '--no-zygote',
    '--disable-gpu'
  ]
};
