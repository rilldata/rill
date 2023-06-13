import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

export default defineConfig(() => {
  const config = {
    resolve: {
      alias: {
        src: "/src", // trick to get absolute imports to work
        "@rilldata/web-local": "/src",
        "@rilldata/web-common": "/../web-common/src",
      },
    },
    plugins: [svelte()],
    test: {
      coverage: {
        provider: "c8",
        src: ["./src"],
        all: true,
      },
    },
  };

  if (process.env["STORYBOOK_MODE"] === "true") {
    console.log(
      "STORYBOOK_MODE===true, updating SvelteKit $app/environment alias"
    );
    config.resolve.alias["$app/environment"] =
      "/../web-common/.storybook/app-environment.mock.ts";
  }

  return config;
});
