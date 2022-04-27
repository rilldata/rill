FROM rilldata/duckdb:0.3.4

WORKDIR /app

COPY package.json package-lock.json \
     tsconfig.json tsconfig.node.json tsconfig.build.json \
     svelte.config.js tailwind.config.cjs postcss.config.cjs .babelrc ./

COPY build-tools build-tools/
COPY src src/

RUN echo "Installing npm dependencies..." && \
    npm install -d

COPY static static/
RUN echo "Compiling the code..." && \
    npm run build

RUN echo "CommonJS vodoo" && \
    /app/build-tools/replace_package_type.sh module commonjs

RUN echo '#!/bin/bash\nnode dist/cli/data-modeler-cli.js "$@"' > /usr/bin/rill-developer && \
    chmod +x /usr/bin/rill-developer

COPY scripts/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
