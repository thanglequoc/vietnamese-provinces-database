export interface ScrapingStats {
  totalProvinces: number;
  processedProvinces: number;
  totalWards: number;
  processedWards: number;
  apiResponsesCaptured: number;
  errors: ScrapingError[];
  startTime: number;
  endTime?: number;
  duration?: number;
}

export interface ScrapingError {
  type: 'province' | 'ward' | 'api' | 'browser' | 'general';
  message: string;
  timestamp: number;
  context?: any;
}
