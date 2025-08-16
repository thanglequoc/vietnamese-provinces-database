
export const SCRAPER_CONFIG = {
  BASE_URL: 'https://sapnhap.bando.com.vn/',
  SELECTORS: {
    PROVINCE_TABLE: '#bangtinh .tabulator-table[role="rowgroup"]',
    PROVINCE_ROW: '.tabulator-row[role="row"]',
    WARD_TABLE: '#bangxa .tabulator-table[role="rowgroup"]',
    WARD_ROW: '.tabulator-row[role="row"]',
    CELL: '.tabulator-cell[role="gridcell"]',
    // Tabulator specific selectors
    PROVINCE_TABLE_HOLDER: '#bangtinh .tabulator-tableholder',
    WARD_TABLE_HOLDER: '#bangxa .tabulator-tableholder',
    TABULATOR_ROOT: '.tabulator'
  },
  TIMEOUTS: {
    PAGE_LOAD: 30000,
    ELEMENT_WAIT: 10000,
    REQUEST_WAIT: 5000,
    BETWEEN_CLICKS: 1000,
    SCROLL_WAIT: 2000,
    STABILIZATION_WAIT: 1000
  },
  RETRY: {
    MAX_ATTEMPTS: 3,
    DELAY: 2000
  },
  SCROLLING: {
    SCROLL_STEP: 200, // Smaller steps for Tabulator virtual scrolling
    MAX_SCROLL_ATTEMPTS: 50, // Reduced for virtual tables
    SCROLL_DELAY: 300, // Faster response for virtual scrolling
    STABILITY_CHECK_ATTEMPTS: 3,
    DUPLICATE_CHECK_THRESHOLD: 3, // Lower threshold for virtual scrolling
    VIRTUAL_SCROLL_WAIT: 200, // Wait for virtual scroll updates
    ESTIMATED_ROW_HEIGHT: 29 // Based on your HTML structure
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
