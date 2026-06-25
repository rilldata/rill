import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";

/**
 * Captures the embed's `console.log` messages (the postMessage protocol echoes
 * its traffic to the console) and exposes web-first assertions over them. The
 * captured messages are encapsulated here rather than passed around as a shared
 * array, so it is clear that the listener owns and mutates them.
 */
class EmbedMessageRecorder {
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

test.describe("Embeds", () => {
  test.describe("embedded explore", () => {
    test("embeds should load", async ({ embedPage }) => {
      const frame = embedPage.frameLocator("iframe");

      await expect(
        frame.getByRole("button", {
          name: "Advertising Spend Overall $20,603",
        }),
      ).toBeVisible();
    });

    test("state is emitted for embeds", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await frame.getByRole("row", { name: "Instacart $2.1k" }).click();

      await recorder.expectContaining(
        "tr=P7D&grain=day&f=advertiser_name+IN+%28%27Instacart%27%29",
      );
    });

    test("reports is disabled because of embed only feature flag", async ({
      embedPage,
    }) => {
      const frame = embedPage.frameLocator("iframe");

      // Open Adomain dimenions table.
      await frame
        .getByLabel("Open dimension details")
        .filter({ hasText: "Adomain" })
        .click();

      // Click export button
      await frame.getByLabel("Export dimension table data").click();
      // Export as csv is available.
      await expect(
        frame.getByRole("menuitem", { name: "Export as CSV" }),
      ).toBeVisible();
      // Create schedule report is not.
      await expect(
        frame.getByRole("menuitem", { name: "Create scheduled report..." }),
      ).not.toBeVisible();
    });

    test("getState returns from embed", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await frame.getByRole("row", { name: "Instacart $2.1k" }).click();
      // Wait for the click's state change to propagate before requesting it.
      await recorder.expectContaining(
        "f=advertiser_name+IN+%28%27Instacart%27%29",
      );

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 1337, method: "getState" },
          "*",
        );
      });

      await recorder.expectContaining(
        `{"id":1337,"result":{"state":"tr=P7D&grain=day&f=advertiser_name+IN+%28%27Instacart%27%29"}}`,
      );
    });

    test("setState changes embedded explore", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1337,
            method: "setState",
            params:
              "tr=P7D&grain=day&f=advertiser_name+IN+%28%27Instacart%27%29",
          },
          "*",
        );
      });
      await expect(
        frame.getByRole("row", { name: "Instacart $2.1k" }),
      ).toBeVisible();
      await recorder.expectContaining(`{"id":1337,"result":true}`);

      // Set new rill syntax that includes `+` in the syntax.
      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1338,
            method: "setState",
            params:
              "tr=2D+as+of+latest%2FD%2B1D&grain=day&f=advertiser_name+IN+%28%27Instacart%27%29",
          },
          "*",
        );
      });
      await expect(
        frame.getByRole("row", { name: "Instacart $1.1k" }),
      ).toBeVisible();
      await recorder.expectContaining(`{"id":1338,"result":true}`);
    });

    test("getThemeMode returns current theme mode", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 2001, method: "getThemeMode" },
          "*",
        );
      });

      await recorder.expectMatching(
        (msg) =>
          msg.includes(`"method":"getThemeMode"`) ||
          (msg.includes(`"id":2001`) &&
            (msg.includes(`"themeMode":"light"`) ||
              msg.includes(`"themeMode":"dark"`) ||
              msg.includes(`"themeMode":"system"`))),
        'message with "method":"getThemeMode" OR id:2001 with themeMode light/dark/system',
      );
    });

    test("setThemeMode changes theme to dark", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 2002,
            method: "setThemeMode",
            params: "dark",
          },
          "*",
        );
      });

      // Check that dark class is applied to the document
      await expect(frame.locator("html.dark")).toBeAttached();
      await recorder.expectContaining(`{"id":2002,"result":true}`);
    });

    test("setThemeMode changes theme to light", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      // First set to dark
      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 2003,
            method: "setThemeMode",
            params: "dark",
          },
          "*",
        );
      });

      await expect(frame.locator("html.dark")).toBeAttached();

      // Then set to light
      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 2004,
            method: "setThemeMode",
            params: "light",
          },
          "*",
        );
      });

      // Check that dark class is not present
      await expect(frame.locator("html.dark")).not.toBeAttached();
      await recorder.expectContaining(`{"id":2004,"result":true}`);
    });

    test("setThemeMode changes theme to system", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 2005,
            method: "setThemeMode",
            params: "system",
          },
          "*",
        );
      });

      await recorder.expectContaining(`{"id":2005,"result":true}`);
    });

    test("setThemeMode rejects invalid theme mode", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 2006,
            method: "setThemeMode",
            params: "invalid",
          },
          "*",
        );
      });

      await recorder.expectMatching(
        (msg) =>
          msg.includes(`"id":2006`) &&
          msg.includes(`"error"`) &&
          msg.includes(`themeMode`),
        'message with id:2006, "error", and "themeMode" (error response for invalid theme)',
      );
    });

    test.describe("embedded explore with initial state", () => {
      test.use({
        embeddedInitialState:
          "&tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
      });

      test("init state is applied to dashboard", async ({ embedPage }) => {
        const recorder = new EmbedMessageRecorder(embedPage);
        await recorder.waitForReady();
        const frame = embedPage.frameLocator("iframe");

        await expect(
          frame.getByRole("button", {
            name: "Advertising Spend Overall $252.33",
          }),
        ).toContainText(
          /Advertising Spend Overall\s+\$252.33\s+-\$52.08\s+-17%/m,
        );

        await recorder.expectContaining(
          "tr=PT6H&compare_tr=rill-PP&grain=hour&f=advertiser_name+IN+%28%27Instacart%27%29",
        );
      });
    });
  });

  test.describe("embedded canvas", () => {
    test.use({
      embeddedResourceName: "bids_canvas",
      embeddedResourceType: "rill.runtime.v1.Canvas",
    });

    test("embeds should load", async ({ embedPage }) => {
      const frame = embedPage.frameLocator("iframe");

      await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
        /Advertising Spend Overall\s+\$3,900\s+\+\$1,858 \+91%\s+vs previous period/m,
      );
    });

    test("state is emitted for embeds", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await frame
        .getByRole("row", { name: "Instacart $1.1k" })
        .scrollIntoViewIfNeeded();
      await frame.getByRole("row", { name: "Instacart $1.1k" }).click();

      await recorder.expectContaining(
        "tr=PT24H&compare_tr=rill-PP&f.bids_metrics=advertiser_name+IN+%28%27Instacart%27%29",
      );
    });

    test("getState returns from embed", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await frame
        .getByRole("row", { name: "Instacart $1.1k" })
        .scrollIntoViewIfNeeded();
      await frame.getByRole("row", { name: "Instacart $1.1k" }).click();
      // Wait for the click's state change to propagate before requesting it.
      await recorder.expectContaining(
        "f.bids_metrics=advertiser_name+IN+%28%27Instacart%27%29",
      );

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 1337, method: "getState" },
          "*",
        );
      });

      await recorder.expectContaining(
        `{"id":1337,"result":{"state":"tr=PT24H&compare_tr=rill-PP&f.bids_metrics=advertiser_name+IN+%28%27Instacart%27%29"}}`,
      );
    });

    test("setState changes embedded canvas", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1337,
            method: "setState",
            params:
              "tr=P7D&compare_tr=rill-PW&f.bids_metrics=advertiser_name+IN+%28%27Instacart%27%29",
          },
          "*",
        );
      });
      await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
        /Advertising Spend Overall\s*\$2,066\s*\+\$1,926 \+1k%\s*vs previous week/,
      );
      await recorder.expectContaining(`{"id":1337,"result":true}`);

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1338,
            method: "setState",
            params:
              "tr=2D+as+of+latest%2FD%2B1D&grain=day&compare_tr=rill-PP&f.bids_metrics=advertiser_name+IN+%28%27Instacart%27%29",
          },
          "*",
        );
      });
      await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
        /Advertising Spend Overall\s*\$1,128\s*\+\$1,075 \+2k%\s*vs previous period/,
      );
      await recorder.expectContaining(`{"id":1338,"result":true}`);
    });

    test("getThemeMode returns current theme mode for canvas", async ({
      embedPage,
    }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 3001, method: "getThemeMode" },
          "*",
        );
      });

      await recorder.expectMatching(
        (msg) =>
          msg.includes(`"id":3001`) &&
          (msg.includes(`"themeMode":"light"`) ||
            msg.includes(`"themeMode":"dark"`) ||
            msg.includes(`"themeMode":"system"`)),
        "message with id:3001 and themeMode light/dark/system",
      );
    });

    test("setThemeMode works for canvas", async ({ embedPage }) => {
      const recorder = new EmbedMessageRecorder(embedPage);
      await recorder.waitForReady();
      const frame = embedPage.frameLocator("iframe");

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 3002,
            method: "setThemeMode",
            params: "dark",
          },
          "*",
        );
      });

      await expect(frame.locator("html.dark")).toBeAttached();
      await recorder.expectContaining(`{"id":3002,"result":true}`);
    });

    test.describe("embedded canvas with initial state", () => {
      test.use({
        embeddedInitialState:
          "&tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
      });

      test("init state is applied to canvas", async ({ embedPage }) => {
        const recorder = new EmbedMessageRecorder(embedPage);
        await recorder.waitForReady();
        const frame = embedPage.frameLocator("iframe");

        await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
          /Advertising Spend Overall\s+\$252.33\s+-\$52.08 -17%\s+vs previous period/m,
        );

        await recorder.expectContaining(
          "tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+%28%27Instacart%27%29",
        );
      });
    });
  });

  test("navigation works as expected", async ({ embedPage }) => {
    const recorder = new EmbedMessageRecorder(embedPage);
    await recorder.waitForReady();
    const frame = embedPage.frameLocator("iframe");

    // Time range is the default
    await expect(frame.getByText("Last 7 days")).toBeVisible();

    // Select "Last 14 Days" as time range
    // Open the menu
    // (Note we cannot use the `interactWithTimeRangeMenu` helper here since its interface is to check the full page)
    await frame.getByLabel("Select time range").click();
    await frame.getByRole("menuitem", { name: "Last 14 days" }).click();
    // Wait for menu to close
    await expect(
      frame.getByRole("menu", { name: "Select time range" }),
    ).not.toBeVisible();

    // Go to the `Programmatic Ads Auction` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitemcheckbox", {
        name: "Programmatic Ads Auction",
        exact: true,
      })
      .click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"bids_explore","to":"auction_explore"}}`,
    );
    // Time range is still the default
    await expect(frame.getByText("Last 7 days")).toBeVisible();

    // Go to the `Bids Canvas Dashboard` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitemcheckbox", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"auction_explore","to":"bids_canvas"}}`,
    );
    // Time range is still the default
    await expect(frame.getByText("Last 24 hours")).toBeVisible();

    // Select "Last 7 days" as time range
    // Open the menu
    // (Note we cannot use the `interactWithTimeRangeMenu` helper here since its interface is to check the full page)
    await frame.getByLabel("Select time range").click();
    await frame.getByRole("menuitem", { name: "Last 7 days" }).click();
    // Wait for menu to close
    await expect(
      frame.getByRole("menu", { name: "Select time range" }),
    ).not.toBeVisible();

    // Go back to the `Programmatic Ads Bids` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitemcheckbox", { name: "Programmatic Ads Bids" })
      .click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"bids_canvas","to":"bids_explore"}}`,
    );
    // Old selection has persisted
    await expect(frame.getByText("Last 14 days")).toBeVisible();

    // Go back to the `Bids Canvas Dashboard` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitemcheckbox", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"bids_explore","to":"bids_canvas"}}`,
    );

    // Old selection has persisted
    await expect(frame.getByText("Last 7 days")).toBeVisible();

    // Go to `Home` using the breadcrumbs
    await frame.getByText("Home").click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"bids_canvas","to":"dashboardListing"}}`,
    );
    // Check that the dashboards are listed
    await expect(
      frame.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible();
    await expect(
      frame.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible();

    // Go to `Programmatic Ads Auction` using the links on home
    await frame.getByRole("link", { name: "Programmatic Ads Bids" }).click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"dashboardListing","to":"bids_explore"}}`,
    );
    // Old selection has persisted
    await expect(frame.getByText("Last 14 Days")).toBeVisible();

    // Go to `Home` using the breadcrumbs
    await frame.getByText("Home").click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"bids_explore","to":"dashboardListing"}}`,
    );
    // Go to `Bids Canvas Dashboard` using the links on home
    await frame
      .getByRole("link", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    await recorder.expectContaining(
      `{"method":"navigation","params":{"from":"dashboardListing","to":"bids_canvas"}}`,
    );
    // Old selection has persisted
    await expect(frame.getByText("Last 7 days")).toBeVisible();
  });
});
