import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";

/**
 * Stores embed params in-memory so that the components that manipulate url need not be aware of these.
 * It can also increase the url size unnessarily, especially the `access_token`.
 */
export class EmbedStore {
  public readonly instanceId: string;
  public readonly runtimeHost: string;
  public readonly accessToken: string;
  /**
   * Array of missing required params.
   * instance_id, runtime_host and access_token are required params.
   */
  public readonly missingRequireParams: string[] = [];
  public readonly navigationEnabled: boolean;

  /**
   * Clean session storage for dashboards that are navigated to for the 1st time.
   * This way once the page is loaded, the dashboard state is persisted.
   * But the moment the user moves away to another page within the parent page, then it will be cleared next time the user comes back to the dashboard.
   */
  public readonly visibleExplores = new Set<string>();

  private static _instance: EmbedStore | null = null;

  public static init(url: URL) {
    this._instance = new EmbedStore(url);
  }

  public static getInstance() {
    return this._instance;
  }

  private constructor(url: URL) {
    this.instanceId = url.searchParams.get("instance_id") ?? "";
    this.runtimeHost = url.searchParams.get("runtime_host") ?? "";
    this.accessToken = url.searchParams.get("access_token") ?? "";
    this.navigationEnabled = url.searchParams.get("navigation") === "true";

    if (!this.instanceId) {
      this.missingRequireParams.push("instance_id");
    }
    if (!this.runtimeHost) {
      this.missingRequireParams.push("runtime_host");
    }
    if (!this.accessToken) {
      this.missingRequireParams.push("access_token");
    }

    dynamicHeight.set(url.searchParams.get("dynamic_height") === "true");
  }
}
