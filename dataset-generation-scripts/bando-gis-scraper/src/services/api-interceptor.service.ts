import { Page, Route } from "@playwright/test";
import { APIInterceptedRequest, ResponseType } from "../interfaces";

export class APIInterceptorService {

  private readonly MAX_RETRIES = 5;
  private readonly RETRY_DELAY = 2000; // 2 seconds delay between retries

  private interceptedRequests: APIInterceptedRequest[] = [];
  private currentUserAction: 'province_click' | 'ward_click' | 'page_load' = 'page_load'

  // Categorized response storage
  private responsesByType: Map<ResponseType, APIInterceptedRequest[]> = new Map([
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
        //console.log("Intercepted URL: ", url); // TODO @thangle: DEBUGGING
        const interceptedRequest: APIInterceptedRequest = {
          url: url,
          method: request.method(),
          headers: request.headers(),
          postData: request.postData() || undefined,
          timestamp: Date.now()
        }

        const success = await this.handleRequestWithRetry(route, interceptedRequest);

        if (!success) {
          console.error(`Failed to intercept request after ${this.MAX_RETRIES} retries: ${url}`);
          await route.continue();
        }
      } else {
        await route.continue();
      }
    })
  }

  private async handleRequestWithRetry(route: Route, interceptedRequest: APIInterceptedRequest): Promise<boolean> {
    for (let attempt = 1; attempt <= this.MAX_RETRIES; attempt++) {
      try {
        // console.log(`Attempt ${attempt}/${this.MAX_RETRIES} for URL: ${interceptedRequest.url}`);

        const response = await route.fetch();
        const responseData = await response.json();
        const responseType = this.classifyResponseType(interceptedRequest.url, responseData);

        interceptedRequest.response = responseData;
        interceptedRequest.responseType = responseType;

        this.interceptedRequests.push(interceptedRequest);

        // also store response by types
        this.responsesByType.get(responseType)?.push({
          ...interceptedRequest,
          responseType: responseType
        });

        await route.fulfill({
          response: response,
          body: JSON.stringify(responseData)
        });

        // console.log(`Successfully intercepted request on attempt ${attempt}: ${interceptedRequest.url}`);
        return true;

      } catch (error) {
        console.error(`Error intercepting request (attempt ${attempt}/${this.MAX_RETRIES}):`, error);

        // If this is not the last attempt, wait before retrying
        if (attempt < this.MAX_RETRIES) {
          console.log(`Retrying in ${this.RETRY_DELAY}ms...`);
          await this.delay(this.RETRY_DELAY);
        }
      }
    }
    return false; // All retries failed
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }


  /*
  Decide whether to intercept and track the request
  */
  private shouldInterceptRequest(url: string): boolean {
    const gisServerEndpoint = 'email.bando.com.vn/cgi-bin/qgis_mapserv.fcgi.exe'
    //const gisServerJSONEndpoint = 'https://sapnhap.bando.com.vn/pread_json'
    if (url.includes(gisServerEndpoint) && url.includes('INFO_FORMAT=application%2Fjson')) {
      return true;
    }

    // if (url.includes(gisServerJSONEndpoint)) {
    //   return true;
    // }
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
  getResponsesSince(timestamp: number, type?: ResponseType): APIInterceptedRequest[] {
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
      if (responseDataJsonString.includes('"matinhxa"')) {
        return ResponseType.WARD_GIS
      } else {
        return ResponseType.PROVINCE_GIS
      }
    }
    return ResponseType.UNKNOWN;
  }
}
