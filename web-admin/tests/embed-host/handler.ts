import { RILL_EMBED_SERVICE_TOKEN_FILE } from "@rilldata/web-integration/tests/constants.ts";
import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  EMBED_ADMIN_URL,
  EMBED_DEFAULT_RESOURCE,
  EMBED_DEFAULT_RESOURCE_TYPE,
  EMBED_ORG_NAME,
  EMBED_PROJECT_NAME,
} from "./constants.ts";
import { getEmbedUsers } from "./users.ts";

export type EmbedHostRequest = {
  method: string;
  path: string;
  body: string;
};

export type EmbedHostResponse = {
  status: number;
  contentType: string;
  body: string;
};

type IFrameRequestBody = {
  user?: string;
  navigation?: boolean;
  resource?: string | null;
  open?: string | null;
};

const HOST_DIR = path.dirname(fileURLToPath(import.meta.url));
// web-admin, which the service token file path is relative to.
const WEB_ADMIN_DIR = path.resolve(HOST_DIR, "../../");

// Only these files are served; request paths are never joined to file paths.
const STATIC_FILES: Record<string, { file: string; contentType: string }> = {
  "/": { file: "index.html", contentType: "text/html" },
  "/index.html": { file: "index.html", contentType: "text/html" },
  "/open.html": { file: "open.html", contentType: "text/html" },
  "/app.js": { file: "app.js", contentType: "text/javascript" },
};

/**
 * Handles a request to the embed host.
 * It's shared by the Playwright route interception used in tests and the dev server in `server.ts`, so they can't drift.
 * Like a real embedder's backend, it keeps the service token on the server and only returns iframe URLs to the browser.
 */
export async function handleEmbedHostRequest({
  method,
  path: reqPath,
  body,
}: EmbedHostRequest): Promise<EmbedHostResponse> {
  const staticFile = STATIC_FILES[reqPath];
  if (method === "GET" && staticFile) {
    return {
      status: 200,
      contentType: staticFile.contentType,
      body: await fs.readFile(path.join(HOST_DIR, staticFile.file), "utf-8"),
    };
  }

  if (method === "GET" && reqPath === "/users.json") {
    return json(
      200,
      getEmbedUsers().map(({ key, label }) => ({ key, label })),
    );
  }

  if (method === "POST" && reqPath === "/api/iframe") {
    return getIFrame(JSON.parse(body || "{}") as IFrameRequestBody);
  }

  return { status: 404, contentType: "text/plain", body: "Not found" };
}

async function getIFrame({
  user: userKey = "none",
  navigation = true,
  resource,
  open,
}: IFrameRequestBody): Promise<EmbedHostResponse> {
  const user = getEmbedUsers().find((u) => u.key === userKey);
  if (!user) return json(400, { message: `Unknown user "${userKey}"` });

  const payload: Record<string, unknown> = { navigation };
  if (user.email) {
    payload.external_user_id = user.email;
    if (user.passUserEmail) payload.user_email = user.email;
  }
  if (user.attributes) payload.attributes = user.attributes;
  // The admin service resolves the dashboard from `open`, and rejects it combined with `resource`.
  if (open) {
    payload.open = open;
  } else {
    payload.resource = resource || EMBED_DEFAULT_RESOURCE;
    payload.type = EMBED_DEFAULT_RESOURCE_TYPE;
  }

  const serviceToken = (
    await fs.readFile(
      path.join(WEB_ADMIN_DIR, RILL_EMBED_SERVICE_TOKEN_FILE),
      "utf-8",
    )
  ).trim();
  const resp = await fetch(
    `${EMBED_ADMIN_URL}/v1/orgs/${EMBED_ORG_NAME}/projects/${EMBED_PROJECT_NAME}/iframe`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${serviceToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    },
  );
  // Pass the admin's status and error through unchanged, so tests can assert on them.
  return {
    status: resp.status,
    contentType: "application/json",
    body: await resp.text(),
  };
}

function json(status: number, body: unknown): EmbedHostResponse {
  return {
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  };
}
