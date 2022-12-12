FROM ubuntu:focal

WORKDIR /app
RUN apt-get update && apt-get install -y curl unzip bash libdigest-sha-perl

RUN apt-get update && apt-get install -y curl unzip make g++

RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash -
RUN apt-get install -y nodejs

COPY package.json package-lock.json tsconfig.json ./
COPY web-local/package.json web-local/tsconfig.json web-local/tsconfig.node.json \
    web-local/svelte.config.js web-local/vite.config.ts \
    web-local/tailwind.config.cjs web-local/postcss.config.cjs web-local/.babelrc web-local/
COPY web-common/package.json web-common/orval.config.ts web-common/

COPY build-tools build-tools/
COPY web-local/build-tools web-local/build-tools/
COPY web-local/src web-local/src/
COPY web-common/src web-common/src/

RUN echo "Installing npm dependencies..." && \
    npm install -d

COPY web-local/static web-local/static/
RUN echo "Compiling the code..." && \
    npm run build

RUN echo "CommonJS voodoo" && \
    node build-tools/replace_package_type.cjs module commonjs

RUN echo '#!/bin/bash\nnode web-local/dist/web-local/src/cli/data-modeler-cli.js "$@"' > /usr/bin/rill && \
    chmod +x /usr/bin/rill

COPY scripts/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
