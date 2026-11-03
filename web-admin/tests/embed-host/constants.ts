/**
 * The embed host stands in for a customer's app that embeds Rill dashboards.
 * Tests serve it with Playwright route interception, and devs can serve it with `npm run test:embed-host`.
 * Both use this origin, since `embeds.open_url` in the AdBids project's `rill.yaml` points to it.
 */
export const EMBED_HOST_PORT = 4300;
export const EMBED_HOST_ORIGIN = `http://localhost:${EMBED_HOST_PORT}`;

export const EMBED_ADMIN_URL = "http://localhost:8080";
export const EMBED_ORG_NAME = "e2e";
export const EMBED_PROJECT_NAME = "AdBids";

export const EMBED_DEFAULT_RESOURCE = "AdBids_metrics_explore";
export const EMBED_DEFAULT_RESOURCE_TYPE = "rill.runtime.v1.Explore";
