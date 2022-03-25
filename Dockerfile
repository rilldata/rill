FROM node:16

WORKDIR /app

COPY package.json package-lock.json \
     tsconfig.json tsconfig.build.json tsconfig.node.json \
     svelte.config.js tailwind.config.cjs postcss.config.cjs .babelrc ./
RUN npm install

COPY src src/
COPY static static/
COPY data data/
COPY build-tools build-tools/

RUN npm run build
RUN ./build-tools/replace_package_type.sh module commonjs

EXPOSE 8080/tcp

ENTRYPOINT node dist/cli/data-modeler-cli.js start
