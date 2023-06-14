import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

const alias = {
  src: "/src", // trick to get absolute imports to work
  "@rilldata/web-local": "/src",
  "@rilldata/web-common": "/../web-common/src",
};

if (process.env["STORYBOOK_MODE"] === "true") {
  alias["$app/environment"] =
    "/../web-common/.storybook/app-environment.mock.ts";
}

export default defineConfig({
  resolve: {
    alias,
  },
  plugins: [svelte()],
  test: {
    coverage: {
      provider: "c8",
      src: ["./src"],
      all: true,
    },
  },
});
