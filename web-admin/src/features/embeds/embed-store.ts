import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export class EmbedStore {
  public readonly instanceId: string;
  public readonly runtimeHost: string;
  public readonly accessToken: string;
  public readonly navigationEnabled: boolean;

  private static _instance: EmbedStore | null = null;

  public static init(url: URL) {
    this._instance = new EmbedStore(url);
    const resource = url.searchParams.get("resource");
    if (!resource) {
      return "/-/embed";
    }

    const type =
      url.searchParams.get("type") === ResourceKind.Canvas
        ? "canvas"
        : "explore";
    return `/-/embed/${type}/${resource}`;
  }

  public static getInstance() {
    return this._instance;
  }

  private constructor(url: URL) {
    this.instanceId = url.searchParams.get("instance_id") ?? "";
    this.runtimeHost = url.searchParams.get("runtime_host") ?? "";
    this.accessToken = url.searchParams.get("access_token") ?? "";
    this.navigationEnabled = url.searchParams.get("navigation") === "true";
  }
}
