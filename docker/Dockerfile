ARG PLAYWRIGHT_VERSION

FROM mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION} AS pw-server

WORKDIR /src
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
ENV NVM_DIR=/root/.nvm
RUN bash -c "source $NVM_DIR/nvm.sh && nvm install --lts"

RUN npm install -g bun
RUN bun install -g bunx
RUN bun init -y

ARG PLAYWRIGHT_VERSION
RUN bun install --save -E @playwright/test@${PLAYWRIGHT_VERSION}
RUN bunx -y playwright@${PLAYWRIGHT_VERSION} install --with-deps

COPY proxy.js chromium.js firefox.js webkit.js /src/

ENTRYPOINT [ "/bin/sh", "-c" ]
