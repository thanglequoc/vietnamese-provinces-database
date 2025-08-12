export interface ScrapingStats {
  totalProvinces: number;
  totalWards: number;
  successfulRequests: number;
  failedRequests: number;
  startTime: Date;
  endTime?: Date;
  duration?: number;
}

export interface ScrapingError {
  type: 'province' | 'ward' | 'api' | 'browser' | 'general';
  message: string;
  timestamp: number;
  context?: any;
}
