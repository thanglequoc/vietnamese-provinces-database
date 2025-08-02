
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
