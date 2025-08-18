# GIS Scraper

Scrape the HTTP request on sapnhan.bando.com.vn.
Automation script created with [Playwright](https://playwright.dev).

It will simulate click on every province and every ward row data, and capture the HTTP request to dump to JSON file.

## How to run

Install Node, Playwright on the system.

Install dependency

> yarn install


Run scraping with
> yarn dev

JSON output dumped at the `./output/` folder on this project directory.