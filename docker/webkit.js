import { webkit } from '@playwright/test';

import { verify } from "./proxy.js";
verify();

(async () => {
    const server = await webkit.launchServer({ proxy: { server: process.argv[2] }, headless: true, port: 1024 + 2, wsPath: 'webkit' });
    console.log("ready endpoint:", server.wsEndpoint());
})();