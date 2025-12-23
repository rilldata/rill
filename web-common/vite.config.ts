import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";
import type { Alias } from "vite";
import { svelteTesting } from "@testing-library/svelte/vite";

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
      workspace: [
        {
          extends: "./vite.config.ts",
          plugins: [svelteTesting()],
          test: {
            name: "client",
            environment: "jsdom",
            clearMocks: true,
            setupFiles: ["./vitest-setup.ts"],
            globals: true,
            coverage: {
              provider: "v8",
              src: ["./src"],
              all: true,
            },
          },
        },
      ],
    },
  };
});
