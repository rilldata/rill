import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

const aliases: any = [
  {
    find: "src",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-local",
    replacement: "/src",
  },
  {
    find: "@rilldata/web-common",
    replacement: "/../web-common/src",
  },
];

export default defineConfig(({ mode, command, ssrBuild }) => {
  if (mode === "test") {
    aliases.push({
      find: /^svelte$/,
      replacement: "/../node_modules/svelte/index.mjs",
    });
  }

  return {
    resolve: {
      alias: aliases,
    },
    plugins: [svelte()],
    test: {
      coverage: {
        provider: "c8",
        src: ["./src"],
        all: true,
      },
      environment: "jsdom",
    },
  };
});
