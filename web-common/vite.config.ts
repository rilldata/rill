import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig, type Plugin } from "vitest/config";
import type { Alias } from "vite";

const alias: Alias[] = [
  {
    find: "src",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-common",
    replacement: "/../web-common/src",
  },
];

if (process.env["STORYBOOK_MODE"] === "true") {
  alias.push({
    find: "$app/environment",
    replacement: "/../web-common/.storybook/app-environment.mock.ts",
  });
}

export default defineConfig(({ mode }) => {
  if (mode === "test") {
    alias.push({
      find: "$app/environment",
      replacement: "/../web-common/.storybook/app-environment.mock.ts",
    });
  }

  return {
    resolve: {
      alias,
    },
    // Temporary casting to `Plugin[]` to avoid TS error.
    plugins: [sveltekit() as unknown as Plugin[]],
    test: {
      // This alias fixes `onMount` not getting called during vitest unit tests.
      // See: https://stackoverflow.com/questions/76577665/vitest-and-svelte-components-onmount
      alias: [{ find: /^svelte$/, replacement: "svelte/internal" }],
      coverage: {
        provider: "v8",
        src: ["./src"],
        all: true,
      },
      environment: "jsdom",
    },
  };
});
