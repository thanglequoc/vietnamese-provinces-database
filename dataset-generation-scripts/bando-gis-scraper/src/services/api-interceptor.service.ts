import { Page, Route } from "@playwright/test";
import { APIInterceptedRequest, ResponseType } from "../interfaces";

export class APIInterceptorService {
  private interceptedRequests: APIInterceptedRequest[] = [];
  private currentUserAction: 'province_click' | 'ward_click' | 'page_load' = 'page_load'

  // Categorized response storage
    private responsesByType: Map<ResponseType, any[]> = new Map([
      [ResponseType.PROVINCE_INFO, []],
      [ResponseType.WARD_INFO, []],
      [ResponseType.PROVINCE_GIS, []],
      [ResponseType.WARD_GIS, []],
      [ResponseType.UNKNOWN, []]
    ]);

  async setupInterceptor(page: Page): Promise<void> {
    await page.route('**/*', async (route: Route) => {
      const request = route.request();
      const url = request.url();

      if (this.shouldInterceptRequest(url)) {
        console.log("Intercepted URL: ", url);
        const interceptedRequest: APIInterceptedRequest = {
          url: url,
          method: request.method(),
          headers: request.headers(),
          postData: request.postData() || undefined,
          timestamp: Date.now()
        }
        
        try {
          const response = await route.fetch();
          const responseData = await response.json();
          const responseType = this.classifyResponseType(url, responseData);
          
          interceptedRequest.response = responseData;
          interceptedRequest.responseType = responseType;

          this.interceptedRequests.push(interceptedRequest);

          // also store response by types
          this.responsesByType.get(responseType)?.push({
            ...interceptedRequest,
            responseType: responseType
          })

          
          await route.fulfill({
            response: response,
            body: JSON.stringify(responseData)
          });
        } catch (error) {
          console.error('Error intercepting request:', error);
          await route.continue();
        }
      } else {
        await route.continue();
      }
    })
  }

  private shouldInterceptRequest(url: string): boolean {
    const gisServerEndpoint = 'email.bando.com.vn/cgi-bin/qgis_mapserv.fcgi.exe'
    const gisServerJSONEndpoint = 'https://sapnhap.bando.com.vn/pread_json'
    if (url.includes(gisServerEndpoint) && url.includes('INFO_FORMAT=application%2Fjson')) {
      return true;
    }
    
    if (url.includes(gisServerJSONEndpoint)) {
      return true;
    }
    return false;
  }

  setUserAction(action: 'province_click' | 'ward_click' | 'page_load'): void {
    this.currentUserAction = action;
  }

  // Clear data with optional type filter
  clearInterceptedData(type?: ResponseType): void {
    if (type) {
      this.responsesByType.set(type, []);
      this.interceptedRequests = this.interceptedRequests.filter(req => req.responseType !== type);
    } else {
      this.interceptedRequests = [];
      this.responsesByType.forEach((_, key) => {
        this.responsesByType.set(key, []);
      });
    }
  }

  // Get responses since a specific timestamp (useful for getting responses after a click)
  getResponsesSince(timestamp: number, type?: ResponseType): any[] {
    let responses = this.interceptedRequests.filter(req => req.timestamp > timestamp);
    
    if (type) {
      responses = responses.filter(req => req.responseType === type);
    }
    
    return responses;
  }

  private classifyResponseType(url: string, responseData: any): ResponseType {
    const urlLower = url.toLowerCase();

    const responseDataJsonString = JSON.stringify(responseData);

    // Is GIS Server response
    if (urlLower.includes('email.bando.com.vn/cgi-bin/qgis_mapserv.fcgi.exe')) {
      if (responseDataJsonString.includes('"matinh"')) {
        return ResponseType.PROVINCE_GIS
      } else {
        return ResponseType.WARD_GIS
      }
    }

    return ResponseType.UNKNOWN;
  }
}
