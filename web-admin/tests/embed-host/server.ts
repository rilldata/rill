/**
 * Serves the embed host on EMBED_HOST_ORIGIN, to debug embedded reports manually in a normal browser.
 * Open links in report emails caught by Mailpit point here too.
 * Requires `npm run test:setup` to have run, for the service token and the deployed AdBids project.
 *
 * Usage (from web-admin): npm run test:embed-host
 */
import http from "node:http";
import { EMBED_HOST_ORIGIN, EMBED_HOST_PORT } from "./constants.ts";
import { handleEmbedHostRequest } from "./handler.ts";

const server = http.createServer((req, res) => {
  void handle(req, res);
});

server.listen(EMBED_HOST_PORT, () => {
  console.log(`Embed host running at ${EMBED_HOST_ORIGIN}`);
});

async function handle(req: http.IncomingMessage, res: http.ServerResponse) {
  try {
    const chunks: Buffer[] = [];
    for await (const chunk of req) chunks.push(chunk as Buffer);
    const { status, contentType, body } = await handleEmbedHostRequest({
      method: req.method ?? "GET",
      path: new URL(req.url ?? "/", EMBED_HOST_ORIGIN).pathname,
      body: Buffer.concat(chunks).toString("utf-8"),
    });
    res.writeHead(status, { "Content-Type": contentType });
    res.end(body);
  } catch (err) {
    console.error(err);
    res.writeHead(500, { "Content-Type": "text/plain" });
    res.end(err instanceof Error ? err.message : String(err));
  }
}
