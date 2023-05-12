import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import type { UserConfig } from "vite";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

const config: UserConfig = {
  resolve: {
    alias: {
      "@rilldata/web-auth": "/src",
      "@rilldata/web-admin": "/../web-admin/src",
      "@rilldata/web-common": "/../web-common/src",
      "@rilldata/web-local": "/../web-local/src",
    },
  },
  server: {
    port: 3000,
    strictPort: true,
  },
  plugins: [sveltekit()],
};

export default config;
