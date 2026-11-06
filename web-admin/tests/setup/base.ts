import { rillCloud } from "@rilldata/web-common/tests/fixtures/rill-cloud-fixtures";
import { getEmbedHostConfig } from "../embed-host/config.ts";
import { EmbedHostApp } from "../embed-host/embed-host-app.ts";
import { handleEmbedHostRequest } from "../embed-host/handler.ts";

export const test = rillCloud.extend<{
  /**
   * The embed host app in `tests/embed-host`, which embeds the AdBids project.
   * Requests to it are served with route interception, so nothing needs to listen on its port.
   */
  embedHost: EmbedHostApp;
}>({
  embedHost: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
      // Chrome treats pages fulfilled by route interception as public, since they have no server address,
      // and its Local Network Access checks then block them from framing the Rill frontend on localhost.
      permissions: ["local-network-access"],
    });
    await context.route(
      (url) => url.origin === getEmbedHostConfig().origin,
      async (route) => {
        const request = route.request();
        const { status, contentType, body } = await handleEmbedHostRequest({
          method: request.method(),
          path: new URL(request.url()).pathname,
          body: request.postData() ?? "",
        });
        await route.fulfill({ status, contentType, body });
      },
    );
    const page = await context.newPage();

    await use(new EmbedHostApp(page));

    await context.close();
  },
});
