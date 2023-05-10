import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      src: "/src", // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
    },
  },
  test: {},
});
