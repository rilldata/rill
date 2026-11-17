import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// A path to the runtime's assets API, which has been written by hand in markdown HTML
// to show project images in Rill Developer.
const ASSETS_API_PATH = /^\/?v1\/instances\/[^/]+\/assets\/(.+)$/;

/**
 * Returns the project-relative path of a URL that points at a project asset:
 * a file in one of the project's public paths, such as `public/img/logo.svg`.
 * Returns null for a URL that the browser can load as is:
 * one with a scheme (such as `https:` or `data:`) or a protocol-relative one.
 */
export function getProjectAssetPath(url: string): string | null {
  url = url.trim();
  if (/^([a-z][a-z0-9+.-]*:|\/\/)/i.test(url)) return null;

  const path = ASSETS_API_PATH.exec(url)?.[1] ?? url;
  return path.replace(/^(\.?\/)+/, "") || null;
}

/**
 * Loads project assets for a component as blob URLs.
 * The browser can't load them by itself: the runtime only serves them to requests that carry the user's token,
 * which is also what makes them work on Rill Cloud and in embeds.
 * An asset is kept until `destroy`, so re-rendering the component doesn't fetch it again.
 */
export class ProjectAssetLoader {
  private readonly blobUrls = new Map<string, Promise<string>>();

  constructor(private readonly client: RuntimeClient) {}

  /** Returns a URL that an `<img>` can load, given an image URL from a component's spec. */
  resolve(url: string): Promise<string> {
    const path = getProjectAssetPath(url);
    if (path === null) return Promise.resolve(url);

    let blobUrl = this.blobUrls.get(path);
    if (!blobUrl) {
      blobUrl = this.fetchAsBlobUrl(path);
      this.blobUrls.set(path, blobUrl);
      // Forget a failure, so the next render tries again.
      blobUrl.catch(() => this.blobUrls.delete(path));
    }
    return blobUrl;
  }

  /** Releases the loaded assets. */
  destroy() {
    for (const blobUrl of this.blobUrls.values()) {
      blobUrl.then(
        (url) => URL.revokeObjectURL(url),
        () => {},
      );
    }
    this.blobUrls.clear();
  }

  private async fetchAsBlobUrl(path: string): Promise<string> {
    const { host, instanceId } = this.client;
    const headers: Record<string, string> = {};
    const jwt = await this.client.getFreshJwt();
    if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

    const response = await fetch(
      `${host}/v1/instances/${instanceId}/assets/${path}`,
      { headers },
    );
    if (!response.ok) {
      throw new Error(`Failed to load ${path}: HTTP ${response.status}`);
    }
    return URL.createObjectURL(await response.blob());
  }
}
