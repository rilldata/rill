FROM node:16

WORKDIR /app

COPY package.json package-lock.json tsconfig.json tsconfig.node.json ./
RUN npm install

COPY tsconfig.build.json svelte.config.js tailwind.config.cjs postcss.config.cjs .babelrc ./
COPY src src/
COPY static static/
COPY data data/
COPY build-tools build-tools/

RUN npm run build
RUN ./build-tools/replace_package_type.sh module commonjs
RUN echo 'alias rill-developer="node dist/cli/data-modeler-cli.js"' >> ~/.bashrc

EXPOSE 8080/tcp

ENTRYPOINT node dist/cli/data-modeler-cli.js start
