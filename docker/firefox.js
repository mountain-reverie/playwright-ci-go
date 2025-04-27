import { firefox } from '@playwright/test';

import { verify } from "./proxy.js";
verify();

(async () => {
    const server = await firefox.launchServer({ proxy: { server: process.argv[2] }, headless: true, port: 1024 + 1, wsPath: 'firefox' });
    console.log("ready endpoint:", server.wsEndpoint());
})();