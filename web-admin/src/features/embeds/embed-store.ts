export class EmbedStore {
  public readonly instanceId: string;
  public readonly runtimeHost: string;
  public readonly accessToken: string;
  public readonly navigationEnabled: boolean;

  public readonly exploreSeen = new Set<string>();

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
  }
}
