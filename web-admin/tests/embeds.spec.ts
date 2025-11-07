import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";

async function waitForReadyMessage(embedPage: Page, logMessages: string[]) {
  return new Promise<void>((resolve) => {
    embedPage.on("console", async (msg) => {
      if (msg.type() !== "log") return;

      try {
        const args = await Promise.all(
          msg.args().map((arg) => arg.jsonValue()),
        );
        const logMessage = JSON.stringify(args);
        logMessages.push(logMessage);
        if (logMessage.includes(`{"method":"ready"}`)) {
          resolve();
        }
      } catch {
        // Ignore errors in parsing. Any rogue log shouldn't break the test.
        // There is also a race condition when browser/page is closed while we are extracting the values in the await.
      }
    });
  });
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
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await frame.getByRole("row", { name: "Instacart $2.1k" }).click();
      await embedPage.waitForTimeout(500);

      expect(
        logMessages.some((msg) =>
          msg.includes("f=advertiser_name+IN+('Instacart')"),
        ),
      ).toBeTruthy();
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
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await frame.getByRole("row", { name: "Instacart $2.1k" }).click();
      await embedPage.waitForTimeout(500);

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 1337, method: "getState" },
          "*",
        );
      });

      await embedPage.waitForTimeout(500);

      expect(
        logMessages.some((msg) =>
          msg.includes(
            `{"id":1337,"result":{"state":"tr=P7D&grain=day&f=advertiser_name+IN+('Instacart')"}}`,
          ),
        ),
      ).toBeTruthy();
    });

    test("setState changes embedded explore", async ({ embedPage }) => {
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1337,
            method: "setState",
            params: "tr=P7D&grain=day&f=advertiser_name+IN+('Instacart')",
          },
          "*",
        );
      });

      await expect(
        frame.getByRole("row", { name: "Instacart $2.1k" }),
      ).toBeVisible();
      expect(
        logMessages.some((msg) => msg.includes(`{"id":1337,"result":true}`)),
      ).toBeTruthy();
    });

    test.describe("embedded explore with initial state", () => {
      test.use({
        embeddedInitialState:
          "&tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
      });

      test("init state is applied to dashboard", async ({ embedPage }) => {
        const logMessages: string[] = [];
        await waitForReadyMessage(embedPage, logMessages);
        const frame = embedPage.frameLocator("iframe");

        await expect(
          frame.getByRole("button", {
            name: "Advertising Spend Overall $252.33",
          }),
        ).toContainText(
          /Advertising Spend Overall\s+\$252.33\s+-\$52.08\s+-17%/m,
        );
        await embedPage.waitForTimeout(500);

        expect(
          logMessages.some((msg) =>
            msg.includes(
              "tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
            ),
          ),
        ).toBeTruthy();
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
        /Advertising Spend Overall\s+\$3,900\s+\+\$1,858 \+91%\s+vs previous day/m,
      );
    });

    test("state is emitted for embeds", async ({ embedPage }) => {
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await frame
        .getByRole("row", { name: "Instacart $1.1k" })
        .scrollIntoViewIfNeeded();
      await frame.getByRole("row", { name: "Instacart $1.1k" }).click();
      await embedPage.waitForTimeout(500);

      expect(
        logMessages.some((msg) =>
          msg.includes(
            "tr=PT24H&compare_tr=rill-PD&f=advertiser_name+IN+('Instacart')",
          ),
        ),
      ).toBeTruthy();
    });

    test("getState returns from embed", async ({ embedPage }) => {
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await frame
        .getByRole("row", { name: "Instacart $1.1k" })
        .scrollIntoViewIfNeeded();
      await frame.getByRole("row", { name: "Instacart $1.1k" }).click();
      await embedPage.waitForTimeout(500);

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          { id: 1337, method: "getState" },
          "*",
        );
      });

      await embedPage.waitForTimeout(500);

      expect(
        logMessages.some((msg) =>
          msg.includes(
            `{"id":1337,"result":{"state":"tr=PT24H&compare_tr=rill-PD&f=advertiser_name+IN+('Instacart')"}}`,
          ),
        ),
      ).toBeTruthy();
    });

    test("setState changes embedded canvas", async ({ embedPage }) => {
      const logMessages: string[] = [];
      await waitForReadyMessage(embedPage, logMessages);
      const frame = embedPage.frameLocator("iframe");

      await embedPage.waitForTimeout(500);

      await embedPage.evaluate(() => {
        const iframe = document.querySelector("iframe");
        iframe?.contentWindow?.postMessage(
          {
            id: 1337,
            method: "setState",
            params:
              "tr=P7D&compare_tr=rill-PW&f=advertiser_name+IN+('Instacart')",
          },
          "*",
        );
      });

      await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
        /Advertising Spend Overall\s*\$2,066\s*\+\$1,926 \+1k%\s*vs previous week/,
      );
      expect(
        logMessages.some((msg) => msg.includes(`{"id":1337,"result":true}`)),
      ).toBeTruthy();
    });

    test.describe("embedded canvas with initial state", () => {
      test.use({
        embeddedInitialState:
          "&tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
      });

      test("init state is applied to canvas", async ({ embedPage }) => {
        const logMessages: string[] = [];
        await waitForReadyMessage(embedPage, logMessages);
        const frame = embedPage.frameLocator("iframe");

        await expect(frame.getByLabel("overall_spend KPI data")).toContainText(
          /Advertising Spend Overall\s+\$252.33\s+-\$52.08 -17%\s+vs previous period/m,
        );
        await embedPage.waitForTimeout(500);

        expect(
          logMessages.some((msg) =>
            msg.includes(
              "tr=PT6H&compare_tr=rill-PP&f=advertiser_name+IN+('Instacart')",
            ),
          ),
        ).toBeTruthy();
      });
    });
  });

  test("navigation works as expected", async ({ embedPage }) => {
    const logMessages: string[] = [];
    await waitForReadyMessage(embedPage, logMessages);
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
      .getByRole("menuitem", {
        name: "Programmatic Ads Auction",
        exact: true,
      })
      .click();
    // Time range is still the default
    await expect(frame.getByText("Last 7 days")).toBeVisible();

    // Go to the `Bids Canvas Dashboard` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitem", { name: "Bids Canvas Dashboard" })
      .first()
      .click();
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
      .getByRole("menuitem", { name: "Programmatic Ads Bids" })
      .click();
    // Old selection has persisted
    await expect(frame.getByText("Last 14 days")).toBeVisible();

    // Go back to the `Bids Canvas Dashboard` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitem", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    // Old selection has persisted
    await expect(frame.getByText("Last 7 days")).toBeVisible();

    // Go to `Home` using the breadcrumbs
    await frame.getByText("Home").click();
    // Check that the dashboards are listed
    await expect(
      frame.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible();
    await expect(
      frame.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible();

    // Go to `Programmatic Ads Auction` using the links on home
    await frame.getByRole("link", { name: "Programmatic Ads Bids" }).click();
    // Old selection has persisted
    await expect(frame.getByText("Last 14 Days")).toBeVisible();

    // Go to `Home` using the breadcrumbs
    await frame.getByText("Home").click();
    // Go to `Bids Canvas Dashboard` using the links on home
    await frame
      .getByRole("link", { name: "Bids Canvas Dashboard" })
      .first()
      .click();
    // Old selection has persisted
    await expect(frame.getByText("Last 7 days")).toBeVisible();
  });
});
