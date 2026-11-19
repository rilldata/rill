/**
 * Serves the embed host, to debug embedded reports manually in a normal browser.
 * Open links in report emails caught by Mailpit point here too.
 * Requires `npm run test:setup` to have run, for the service token and the deployed AdBids project,
 * unless `tests/embed-host/.env.test` points it elsewhere; see `tests/README.md`.
 *
 * Usage (from web-admin): npm run test:embed-host
 */
import http from "node:http";
import { getEmbedHostConfig, loadEmbedHostDevOverrides } from "./config.ts";
import { handleEmbedHostRequest } from "./handler.ts";

loadEmbedHostDevOverrides();
const config = getEmbedHostConfig();

const server = http.createServer((req, res) => {
  void handle(req, res);
});

server.listen(config.port, () => {
  console.log(`Embed host running at ${config.origin}`);
});

async function handle(req: http.IncomingMessage, res: http.ServerResponse) {
  try {
    const chunks: Buffer[] = [];
    for await (const chunk of req) chunks.push(chunk as Buffer);
    const { status, contentType, body } = await handleEmbedHostRequest({
      method: req.method ?? "GET",
      path: new URL(req.url ?? "/", config.origin).pathname,
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
