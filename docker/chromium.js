import { chromium } from '@playwright/test';

import { verify } from "./proxy.js";
verify();

(async () => {
    const server = await chromium.launchServer({ proxy: { server: process.argv[2] }, headless: true, host: '0.0.0.0', port: 1024 + 3, wsPath: 'chromium' });
    console.log("ready endpoint:", server.wsEndpoint());
})();