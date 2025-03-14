import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";
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
    plugins: [sveltekit()],
    test: {
      // This alias fixes `onMount` not getting called during vitest unit tests.
      // See: https://stackoverflow.com/questions/76577665/vitest-and-svelte-components-onmount
      alias: [{ find: /^svelte$/, replacement: "svelte/internal" }],
      setupFiles: ["./vitest-setup.js"],
      globals: true,
      coverage: {
        provider: "v8",
        src: ["./src"],
        all: true,
      },
      environment: "jsdom",
    },
  };
});
