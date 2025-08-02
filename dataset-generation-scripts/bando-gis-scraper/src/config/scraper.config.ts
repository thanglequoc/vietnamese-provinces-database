
export const SCRAPER_CONFIG = {
  BASE_URL: 'https://sapnhap.bando.com.vn/',
  SELECTORS: {
    PROVINCE_TABLE: '.tabulator-table[role="rowgroup"]',
    PROVINCE_ROW: '.tabulator-row[role="row"]',
    WARD_TABLE: '.tabulator-table[role="rowgroup"]',
    WARD_ROW: '.tabulator-row[role="row"]',
    CELL: '.tabulator-cell[role="gridcell"]'
  },
  TIMEOUTS: {
    PAGE_LOAD: 30000,
    ELEMENT_WAIT: 10000,
    REQUEST_WAIT: 5000,
    BETWEEN_CLICKS: 1000
  },
  RETRY: {
    MAX_ATTEMPTS: 3,
    DELAY: 2000
  }
};

export const BROWSER_OPTIONS = {
  slowMo: 50,
  args: [
    '--no-sandbox',
    '--disable-setuid-sandbox',
    '--disable-dev-shm-usage',
    '--disable-accelerated-2d-canvas',
    '--no-first-run',
    '--no-zygote',
    '--disable-gpu'
  ]
};
