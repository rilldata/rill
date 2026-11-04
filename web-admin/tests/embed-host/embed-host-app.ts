import {
  expect,
  type ConsoleMessage,
  type FrameLocator,
  type Page,
} from "@playwright/test";
import { EmbedMessageRecorder } from "@rilldata/web-common/tests/utils/embed-message-recorder.ts";
import { getEmbedHostConfig } from "./config.ts";
import { getEmbedUser, type EmbedUserKey } from "./users.ts";

export type EmbedHostPage = "home" | "reports" | "alerts";

/**
 * Drives the embed host page (index.html and open.html) in tests.
 */
export class EmbedHostApp {
  public readonly frame: FrameLocator;
  /** Attached before the first navigation, so it captures the embed's `ready` message. */
  public readonly recorder: EmbedMessageRecorder;
  // Resolves when the iframe loaded by the latest navigation posts its `ready` message.
  private iframeReady: Promise<unknown> = Promise.resolve();

  public constructor(public readonly page: Page) {
    this.frame = page.frameLocator("#rill-frame");
    this.recorder = new EmbedMessageRecorder(page);
  }

  public async goto({
    user,
    navigation = true,
    resource,
  }: {
    user: EmbedUserKey;
    navigation?: boolean;
    resource?: string;
  }) {
    const url = new URL("/", getEmbedHostConfig().origin);
    url.searchParams.set("user", user);
    url.searchParams.set("navigation", String(navigation));
    if (resource) url.searchParams.set("resource", resource);
    this.expectIframeLoad();
    await this.page.goto(url.toString());
  }

  /** Goes to an open link from a report email, and picks the user on the open page. */
  public async openLink(link: string, user: EmbedUserKey) {
    await this.page.goto(link);
    this.expectIframeLoad();
    await this.page
      .getByRole("link", { name: getEmbedUser(user).label, exact: true })
      .click();
    await this.page.waitForURL(
      (url) => url.pathname === "/" && url.searchParams.get("user") === user,
    );
  }

  /** Toggles the host's navigation checkbox, which reloads the iframe. */
  public async setNavigation(navigation: boolean) {
    this.expectIframeLoad();
    await this.page
      .getByRole("checkbox", { name: "Navigation" })
      .setChecked(navigation);
  }

  /** Clicks one of the host's buttons that post a `navigate` message to the embed. */
  public async navigate(name: EmbedHostPage) {
    // Messages posted before the embed is ready are dropped.
    await this.iframeReady;
    const label = name.charAt(0).toUpperCase() + name.slice(1);
    await this.page.getByRole("button", { name: label, exact: true }).click();
  }

  public async expectError(text: string | RegExp) {
    await expect(this.page.getByLabel("Embed error")).toContainText(text);
  }

  private expectIframeLoad() {
    this.iframeReady = this.page.waitForEvent("console", {
      predicate: isReadyMessage,
      timeout: 30_000,
    });
    // Avoid an unhandled rejection if the page closes before the embed is ready; awaiting it still throws.
    this.iframeReady.catch(() => {});
  }
}

async function isReadyMessage(msg: ConsoleMessage) {
  if (msg.type() !== "log") return false;
  try {
    const args = (await Promise.all(
      msg.args().map((arg) => arg.jsonValue()),
    )) as unknown[];
    return args.some(
      (arg) => (arg as { method?: string } | null)?.method === "ready",
    );
  } catch {
    return false;
  }
}
