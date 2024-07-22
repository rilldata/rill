import { sveltekit } from "@sveltejs/kit/vite";
import dns from "dns";
import type { UserConfig } from "vite";
import { readFileSync } from "fs";

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

function readPublicEmailDomains() {
  const contents = readFileSync(
    __dirname + "/../admin/pkg/publicemail/public_email_providers_list",
  ).toString();
  return contents
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => !l.startsWith("#"));
}

export default config;
