import { expect, type Page } from "@playwright/test";

/**
 * Captures the embed's `console.log` messages (the postMessage protocol echoes
 * its traffic to the console) and exposes web-first assertions over them. The
 * captured messages are encapsulated here rather than passed around as a shared
 * array, so it is clear that the listener owns and mutates them.
 */
export class EmbedMessageRecorder {
  private readonly messages: string[] = [];
  private readonly ready: Promise<void>;

  constructor(embedPage: Page) {
    this.ready = new Promise<void>((resolve) => {
      embedPage.on("console", async (msg) => {
        if (msg.type() !== "log") return;

        try {
          const args = await Promise.all(
            msg.args().map((arg) => arg.jsonValue()),
          );
          const logMessage = JSON.stringify(args);
          this.messages.push(logMessage);
          if (logMessage.includes(`{"method":"ready"}`)) {
            resolve();
          }
        } catch {
          // Ignore errors in parsing. Any rogue log shouldn't break the test.
          // There is also a race condition when browser/page is closed while we
          // are extracting the values in the await.
        }
      });
    });
  }

  /** Resolves once the embed has posted its `ready` message. */
  waitForReady() {
    return this.ready;
  }

  /** Asserts that at least one captured message contains the substring. */
  async expectContaining(expectedSubstring: string) {
    const found = await this.poll((msg) => msg.includes(expectedSubstring));
    if (!found) {
      expect
        .soft(
          found,
          `No message containing expected substring.

Expected substring:
  ${expectedSubstring}

Received messages:
${this.formatMessages()}`,
        )
        .toBeTruthy();
    }
  }

  /** Asserts that at least one captured message matches the predicate. */
  async expectMatching(
    predicate: (msg: string) => boolean,
    description: string,
  ) {
    const found = await this.poll(predicate);
    if (!found) {
      expect
        .soft(
          found,
          `No message matching expected condition.

Expected:
  ${description}

Received messages:
${this.formatMessages()}`,
        )
        .toBeTruthy();
    }
  }

  /**
   * Polls the captured messages until one satisfies the predicate, or the
   * timeout elapses. Returns whether a matching message was found, so callers
   * can emit a detailed, debuggable failure rather than a bare timeout.
   */
  private async poll(predicate: (msg: string) => boolean, timeout = 5_000) {
    try {
      await expect
        .poll(() => this.messages.some(predicate), { timeout })
        .toBe(true);
    } catch {
      // Timed out: fall through so the caller can report the captured messages.
    }
    return this.messages.some(predicate);
  }

  private formatMessages() {
    return this.messages.length > 0
      ? this.messages.map((m, i) => `  [${i}]: ${m}`).join("\n")
      : "  (no messages captured)";
  }
}
