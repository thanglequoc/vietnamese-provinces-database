import { ProvinceData, WardData } from "./scraper.interfaces";

export interface ScrapingResult {
  provinces: ProvinceData[];
  wards: WardData[];
  totalRequests: number;
  startTime: Date;
  endTime?: Date;
  duration?: number;
  errors: string[];
}


export interface ScrapingError {
  type: 'province' | 'ward' | 'api' | 'browser' | 'general';
  message: string;
  timestamp: number;
  context?: any;
}
