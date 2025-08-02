import { Page, Route } from "@playwright/test";
import { APIInterceptedRequest } from "../interfaces";

export class APIInterceptorService {
  private interceptedRequests: APIInterceptedRequest[] = [];

  async setupInterceptor(page: Page): Promise<void> {
    await page.route('**/*', async (route: Route) => {
      const request = route.request();
      const url = request.url();

      if (this.isGISRelatedRequest(url)) {
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

          interceptedRequest.response = responseData;
          this.interceptedRequests.push(interceptedRequest);

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

  private isGISRelatedRequest(url: string): boolean {
    const gisRequestEndpointPattern = [
      'email.bando.com.vn/cgi-bin/qgis_mapserv.fcgi.exe',
    ]
    return gisRequestEndpointPattern.some(pattern => url.toLowerCase().includes(pattern.toLowerCase()))
  }
}
