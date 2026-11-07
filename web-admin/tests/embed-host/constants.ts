/**
 * Defaults for the embed host, which stands in for a customer's app that embeds Rill dashboards.
 * Tests always use these; the dev server (`npm run test:embed-host`) lets `tests/embed-host/.env.test` override them.
 * Read them through `getEmbedHostConfig()` in `config.ts`.
 * The port matches `embeds.open_url` in the AdBids project's `rill.yaml`.
 */
export const EMBED_HOST_PORT = 4300;

export const EMBED_ADMIN_URL = "http://localhost:8080";
export const EMBED_ORG_NAME = "e2e";
export const EMBED_PROJECT_NAME = "AdBids";

export const EMBED_DEFAULT_RESOURCE = "AdBids_metrics_explore";
export const EMBED_DEFAULT_RESOURCE_TYPE = "rill.runtime.v1.Explore";
