import { RILL_EMBED_SERVICE_TOKEN_FILE } from "@rilldata/web-integration/tests/constants.ts";
import dotenv from "dotenv";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  EMBED_ADMIN_URL,
  EMBED_DEFAULT_RESOURCE,
  EMBED_DEFAULT_RESOURCE_TYPE,
  EMBED_HOST_PORT,
  EMBED_ORG_NAME,
  EMBED_PROJECT_NAME,
} from "./constants.ts";

export type EmbedHostConfig = {
  port: number;
  origin: string;
  adminUrl: string;
  orgName: string;
  projectName: string;
  defaultResource: string;
  defaultResourceType: string;
  /** The base email the embed users' emails are derived from; see `users.ts`. */
  email: string | undefined;
  /** Reads the service token used to request iframe URLs. */
  getServiceToken: () => Promise<string>;
};

const HOST_DIR = path.dirname(fileURLToPath(import.meta.url));
const WEB_ADMIN_DIR = path.resolve(HOST_DIR, "../../");
const REPO_ROOT = path.resolve(HOST_DIR, "../../../");
const DEV_ENV_FILE = path.join(HOST_DIR, ".env.test");

// Values from `.env.test`; only loaded by the dev server, so tests always use the defaults.
let devOverrides: Record<string, string> = {};
let config: EmbedHostConfig | undefined;

/**
 * Loads overrides from `tests/embed-host/.env.test`, if it exists.
 * Only the dev server in `server.ts` calls this; tests must match the AdBids project deployed by `npm run test:setup`.
 */
export function loadEmbedHostDevOverrides() {
  if (!fs.existsSync(DEV_ENV_FILE)) return;
  devOverrides = dotenv.parse(fs.readFileSync(DEV_ENV_FILE));
  config = undefined;
}

export function getEmbedHostConfig(): EmbedHostConfig {
  if (config) return config;

  // e2e workers don't load the root `.env` (only the setup project does), so load it here for the e2e admin's email.
  // It's loaded lazily, so specs that don't use the embed host don't need it.
  dotenv.config({ path: path.join(REPO_ROOT, ".env") });

  const port = Number(devOverrides.EMBED_HOST_PORT || EMBED_HOST_PORT);
  const serviceToken = devOverrides.EMBED_SERVICE_TOKEN;
  config = {
    port,
    origin: `http://localhost:${port}`,
    adminUrl: devOverrides.EMBED_ADMIN_URL || EMBED_ADMIN_URL,
    orgName: devOverrides.EMBED_ORG_NAME || EMBED_ORG_NAME,
    projectName: devOverrides.EMBED_PROJECT_NAME || EMBED_PROJECT_NAME,
    defaultResource:
      devOverrides.EMBED_DEFAULT_RESOURCE || EMBED_DEFAULT_RESOURCE,
    defaultResourceType:
      devOverrides.EMBED_DEFAULT_RESOURCE_TYPE || EMBED_DEFAULT_RESOURCE_TYPE,
    email:
      devOverrides.EMBED_EMAIL ||
      process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL,
    getServiceToken: async () =>
      serviceToken ||
      (
        await fs.promises.readFile(
          path.join(WEB_ADMIN_DIR, RILL_EMBED_SERVICE_TOKEN_FILE),
          "utf-8",
        )
      ).trim(),
  };
  return config;
}
