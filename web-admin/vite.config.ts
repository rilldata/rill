import { sveltekit } from "@sveltejs/kit/vite";
import type { UserConfig } from "vite";

const config: UserConfig = {
  resolve: {
    alias: {
      "@rilldata/web-admin": "/src",
      "@rilldata/web-common": "/../web-common/src",
      "@rilldata/web-local": "/../web-local/src",
    },
  },
  plugins: [sveltekit()],
};

export default config;
