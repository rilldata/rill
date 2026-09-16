import { createIframeRPCHandler } from "@rilldata/web-common/lib/rpc";
import type { Page } from "@sveltejs/kit";
import { get, writable, type Readable, type Updater } from "svelte/store";

/**
 * Mocking `$app/stores` and `$app/navigation` requires hoisting the shared page
 * store via `vi.hoisted` so it exists before those modules and
 * `init-embed-public-api.ts` are imported. To avoid rearranging imports we hand
 * an empty object to `vi.hoisted` and let this harness attach behaviour to it.
 */
export type HoistedEmbedPage = Readable<Page> & {
  // Mirrors the single `goto(url, { replaceState })` shape used by the embed API.
  goto: (url: URL, opts?: { replaceState?: boolean }) => void;
};

// A message posted from the embed iframe to its parent window: either a JSON-RPC
// response ({ id, result | error }) to a `call`, or a notification ({ method, params }).
type PostedMessage = {
  id?: string | number | null;
  method?: string;
  params?: unknown;
  result?: unknown;
  error?: { code: number; message: string; data?: unknown };
};

export type GotoCall = { url: URL; opts?: { replaceState?: boolean } };

const DEFAULT_URL = "http://localhost/-/embed";
const DEFAULT_ROUTE_ID = "/[organization]/[project]/-/embed";

/**
 * Drives `initEmbedPublicAPI` through the real RPC transport, i.e. as an
 * embedding parent window sees it. It stands in for the parent window
 * (capturing everything the iframe `postMessage`s), installs the real
 * `createIframeRPCHandler` message listener, and dispatches JSON-RPC requests
 * via `window` message events. Only the SvelteKit page store and navigation are
 * mocked at the module level, since those cannot be reached through `window`.
 *
 * Usage:
 * ```
 * const { hoistedPage } = vi.hoisted(() => ({ hoistedPage: {} as HoistedEmbedPage }));
 * vi.mock("$app/navigation", () => ({ goto: (u, o) => hoistedPage.goto(u, o) }));
 * vi.mock("$app/stores", () => ({ page: hoistedPage }));
 *
 * beforeEach(() => {
 *   harness = new EmbedPublicAPIHarness(hoistedPage);
 *   cleanup = initEmbedPublicAPI({} as RuntimeClient); // registers methods, emits "ready"
 * });
 * afterEach(() => {
 *   cleanup();
 *   harness.destroy();
 * });
 * ```
 */
export class EmbedPublicAPIHarness {
  private readonly update: (updater: Updater<Page>) => void;
  // Every `goto` call the embed API makes, in order.
  public readonly gotoCalls: GotoCall[] = [];
  // Every message the iframe posted to its (mocked) parent window, in order.
  public readonly messages: PostedMessage[] = [];

  private readonly removeRPCHandler: () => void;
  private readonly restoreParent: () => void;
  private nextRPCId = 0;
  private readonly pending = new Map<number, (msg: PostedMessage) => void>();

  public constructor(
    private readonly hoistedPage: HoistedEmbedPage,
    initial?: {
      url?: string;
      routeId?: string;
      params?: Record<string, string>;
    },
  ) {
    const { update, subscribe } = writable<Page>({
      url: new URL(initial?.url ?? DEFAULT_URL),
      route: { id: initial?.routeId ?? DEFAULT_ROUTE_ID },
      params: initial?.params ?? {},
    } as Page);
    this.update = update;

    hoistedPage.subscribe = subscribe;
    hoistedPage.goto = (url, opts) => {
      this.gotoCalls.push({ url, opts });
      update((page) => {
        page.url = url;
        return page;
      });
    };

    // Stand in for the parent window. The RPC layer only sends responses and
    // notifications when `window.parent !== window`, and routes them through
    // `window.parent.postMessage`, so capturing that captures the full surface.
    const parentMock = {
      postMessage: (msg: PostedMessage) => {
        this.messages.push(msg);
        if (msg?.id != null) {
          const resolve = this.pending.get(msg.id as number);
          if (resolve) {
            this.pending.delete(msg.id as number);
            resolve(msg);
          }
        }
      },
    };
    const originalParent = Object.getOwnPropertyDescriptor(window, "parent");
    Object.defineProperty(window, "parent", {
      value: parentMock,
      configurable: true,
    });
    this.restoreParent = () => {
      if (originalParent) {
        Object.defineProperty(window, "parent", originalParent);
      } else {
        delete (window as unknown as { parent?: unknown }).parent;
      }
    };

    // Install the real message listener that dispatches to registered methods.
    this.removeRPCHandler = createIframeRPCHandler();
  }

  /**
   * Invoke an RPC method the way a parent window would: post a JSON-RPC request
   * and resolve with the response message ({ id, result } or { id, error }).
   */
  public call(method: string, params?: unknown): Promise<PostedMessage> {
    const id = ++this.nextRPCId;
    const response = new Promise<PostedMessage>((resolve) =>
      this.pending.set(id, resolve),
    );
    window.dispatchEvent(
      new MessageEvent("message", {
        data: { id, method, params },
        source: window,
      }),
    );
    return response;
  }

  /**
   * Put the page store on a given route, e.g. an explore or canvas embed route.
   * Used to exercise route-dependent behaviour such as `setValidState`'s
   * upfront validation, which only applies to explore dashboards.
   */
  public setRoute(routeId: string, params: Record<string, string> = {}) {
    this.update((page) => {
      page.route = { id: routeId } as Page["route"];
      page.params = params;
      return page;
    });
  }

  /**
   * Simulate an external navigation (e.g. the user interacting with the dashboard)
   * that changes only the url's search params. Fires the page store subscribers,
   * which is what the `stateChange` notification listens to.
   */
  public navigateTo(search: string) {
    this.update((page) => {
      const url = new URL(page.url);
      url.search = search;
      page.url = url;
      return page;
    });
  }

  /** Notifications posted so far (messages with a `method`), optionally filtered. */
  public notifications(method?: string): PostedMessage[] {
    const notifications = this.messages.filter((m) => m.method !== undefined);
    if (!method) return notifications;
    return notifications.filter((m) => m.method === method);
  }

  /** Discard all captured messages, e.g. after flushing init-time emissions. */
  public clearMessages() {
    this.messages.length = 0;
  }

  /** The most recent `goto` call, or undefined if none has happened yet. */
  public lastGoto(): GotoCall | undefined {
    return this.gotoCalls.at(-1);
  }

  public get currentUrl(): URL {
    return get(this.hoistedPage).url;
  }

  /** Remove the message listener and restore the real `window.parent`. */
  public destroy() {
    this.removeRPCHandler();
    this.restoreParent();
  }
}
