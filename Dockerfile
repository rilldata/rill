# syntax = docker/dockerfile:1.1-experimental
FROM rilldata/duckdb:0.4.0

WORKDIR /app

COPY package.json package-lock.json \
    tsconfig.json tsconfig.node.json tsconfig.build.json \
    svelte.config.js vite.config.ts tailwind.config.cjs postcss.config.cjs .babelrc ./

COPY build-tools build-tools/
COPY src src/

RUN echo "Installing npm dependencies..." && \
    npm install -d

COPY static static/
RUN echo "Compiling the code..." && \
    npm run build

RUN echo "CommonJS vodoo" && \
    /app/build-tools/replace_package_type.sh module commonjs

RUN echo '#!/bin/bash\nnode dist/cli/data-modeler-cli.js "$@"' > /usr/bin/rill && \
    chmod +x /usr/bin/rill

COPY scripts/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
