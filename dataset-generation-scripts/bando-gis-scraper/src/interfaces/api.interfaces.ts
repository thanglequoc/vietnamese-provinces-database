
// The response model of endpoint
// https://email.bando.com.vn/cgi-bin/qgis_mapserv.fcgi.exe?
export interface CgiBinQGISMapServiceResponse {
  type: string;
  features: {
    type: string;
    id: string;
    geometry: null;
    properties: {
      matinh: string;
    };
  }[];
}

export interface APIInterceptedRequest {
  url: string;
  method: string;
  headers: Record<string, string>;
  postData?: string;
  response?: any;
  timestamp: number;
  responseType?: ResponseType;
  userAction?: 'province_click' | 'ward_click' | 'page_load';
}

export enum ResponseType {
  PROVINCE_INFO = 'province_info',
  WARD_INFO = 'ward_info',
  PROVINCE_GIS = 'province_gis',
  WARD_GIS = 'ward_gis',
  UNKNOWN = 'unknown'
}
