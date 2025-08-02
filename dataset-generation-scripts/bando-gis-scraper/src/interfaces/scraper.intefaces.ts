
export interface ScrapingConfig {
  websiteUrl: string;
  headless?: boolean;
  delayBetweenClicks?: number;
  delayBetweenProvinces?: number;
  outputFile?: string;
  maxRetries?: number;
  timeout?: number;
  userAgent?: string;
  viewport?: {
    width: number;
    height: number;
  };
}

export interface ElementClickResult {
  success: boolean;
  error?: string;
  retryCount: number;
}
