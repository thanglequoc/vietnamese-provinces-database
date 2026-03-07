export interface ProvinceData {
  stt: string;
  ten: string;
  truocsn: string;
  gisServerResponse?: string;
}

export interface WardData {
  stt: string;
  ten: string;
  truocsn: string;
  provinceCode?: string;
  provinceName?: string;
  gisServerResponse?: string;
}

export interface FailedGISItem {
  itemType: 'province' | 'ward';
  itemData: ProvinceData | WardData;
  attempts: number;
  lastError?: string;
  timestamp: Date;
}