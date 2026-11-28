import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
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
 * An asset is kept until its file changes or `destroy` is called, so re-rendering the component doesn't fetch it again.
 */
export class ProjectAssetLoader {
  // Keyed by the asset's path as `splitAssetPath` returns it.
  private readonly blobUrls = new Map<string, Promise<string>>();
  private readonly stopWatching: () => void;
  private destroyed = false;

  /**
   * @param onChange Called when the file of a loaded asset changes or is deleted, such as when it's saved in Rill Developer,
   * so the component can load the asset again.
   */
  constructor(
    private readonly client: RuntimeClient,
    private readonly onChange?: () => void,
  ) {
    this.stopWatching = eventBus.on("project-file-changed", (path) =>
      this.forget(path),
    );
  }

  /** Returns a URL that an `<img>` can load, given an image URL from a component's spec. */
  resolve(url: string): Promise<string> {
    const path = getProjectAssetPath(url);
    return path === null ? Promise.resolve(url) : this.load(path);
  }

  /**
   * Fetches the asset at a project-relative path and returns a blob URL for it.
   * Unlike `resolve`, it never returns its input, so it is safe for a path read from rendered HTML.
   */
  load(path: string): Promise<string> {
    // A blob URL created after `destroy` would never be released.
    // This happens when a component loads an image during teardown.
    if (this.destroyed) {
      return Promise.reject(new Error("The asset loader was destroyed"));
    }

    const segments = splitAssetPath(path);
    const key = segments.join("/");
    let blobUrl = this.blobUrls.get(key);
    if (!blobUrl) {
      const fetched = this.fetchAsBlobUrl(segments);
      this.blobUrls.set(key, fetched);
      // Forget a failure, so the next render tries again.
      fetched.catch(() => {
        if (this.blobUrls.get(key) === fetched) this.blobUrls.delete(key);
      });
      blobUrl = fetched;
    }
    return blobUrl;
  }

  /** Releases the loaded assets. */
  destroy() {
    this.destroyed = true;
    this.stopWatching();
    for (const blobUrl of this.blobUrls.values()) revokeWhenLoaded(blobUrl);
    this.blobUrls.clear();
  }

  // Drops the asset of a changed file, so the next load fetches it again.
  private forget(filePath: string) {
    const key = splitAssetPath(filePath).join("/");
    const blobUrl = this.blobUrls.get(key);
    if (!blobUrl) return;
    this.blobUrls.delete(key);
    revokeWhenLoaded(blobUrl);
    this.onChange?.();
  }

  private async fetchAsBlobUrl(segments: string[]): Promise<string> {
    const response = await fetchProjectAsset(this.client, segments);
    return URL.createObjectURL(await response.blob());
  }
}

/**
 * Fetches a project asset from the runtime's assets API with the user's token,
 * waiting for a token that is about to expire to be refreshed, like queries do.
 * `segments` are the decoded segments of the asset's project-relative path.
 * Rejects when the runtime doesn't return the file.
 */
export async function fetchProjectAsset(
  client: RuntimeClient,
  segments: string[],
): Promise<Response> {
  const headers: Record<string, string> = {};
  const jwt = await client.getFreshJwt();
  if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

  // Encode each segment, so that characters such as `#` and `?` reach the runtime as part of the file name.
  const path = segments.map(encodeURIComponent).join("/");
  const response = await fetch(
    `${client.host}/v1/instances/${client.instanceId}/assets/${path}`,
    { headers },
  );
  if (!response.ok) {
    throw new Error(`Failed to load ${path}: HTTP ${response.status}`);
  }
  return response;
}

/**
 * Splits a project-relative path into decoded segments, without `.` and `..` segments.
 * Paths from markdown can be percent-encoded already: marked writes a space as `%20`.
 * `..` is resolved here rather than by the browser, so that a request never leaves the assets API.
 */
function splitAssetPath(path: string): string[] {
  const segments: string[] = [];
  for (const segment of path.split("/")) {
    let name = segment;
    try {
      name = decodeURIComponent(segment);
    } catch {
      // A `%` that doesn't start an escape sequence is part of the name.
    }
    if (name === "" || name === ".") continue;
    if (name === "..") segments.pop();
    else segments.push(name);
  }
  return segments;
}

function revokeWhenLoaded(blobUrl: Promise<string>) {
  blobUrl.then(
    (url) => URL.revokeObjectURL(url),
    () => {},
  );
}
