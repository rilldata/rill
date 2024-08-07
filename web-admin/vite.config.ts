import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import type { UserConfig } from "vite";
import { readPublicEmailDomains } from "./src/features/projects/user-invite/readPublicEmailDomains";

// print dev server as `localhost` not `127.0.0.1`
dns.setDefaultResultOrder("verbatim");

const config: UserConfig = {
  resolve: {
    alias: {
      "@rilldata/web-admin": "/src",
      "@rilldata/web-common": "/../web-common/src",
    },
  },
  server: {
    port: 3000,
    strictPort: true,
  },
  define: {
    RillPublicEmailDomains: readPublicEmailDomains(),
  },
  plugins: [sveltekit()],
  envDir: "../",
  envPrefix: "RILL_UI_PUBLIC_",
};

export default config;
