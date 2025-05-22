import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";

async function waitForReadyMessage(embedPage: Page, logMessages: string[]) {
  return new Promise<void>((resolve) => {
    embedPage.on("console", async (msg) => {
      if (msg.type() === "log") {
        const args = await Promise.all(
          msg.args().map((arg) => arg.jsonValue()),
        );
        const logMessage = JSON.stringify(args);
        logMessages.push(logMessage);
        if (logMessage.includes(`{"method":"ready"}`)) {
          resolve();
        }
      }
    });
  });
}

test.describe("Embeds", () => {
  test("embeds should load", async ({ embedPage }) => {
    const frame = embedPage.frameLocator("iframe");

    await expect(
      frame.getByRole("button", { name: "Advertising Spend Overall $3,900" }),
    ).toBeVisible();
  });

  test("state is emitted for embeds", async ({ embedPage }) => {
    const logMessages: string[] = [];
    await waitForReadyMessage(embedPage, logMessages);
    const frame = embedPage.frameLocator("iframe");

    await frame.getByRole("row", { name: "Instacart $1.1k" }).click();
    await embedPage.waitForTimeout(500);

    expect(
      logMessages.some((msg) =>
        msg.includes("f=advertiser_name+IN+('Instacart')"),
      ),
    ).toBeTruthy();
  });

  test("getState returns from embed", async ({ embedPage }) => {
    const logMessages: string[] = [];
    await waitForReadyMessage(embedPage, logMessages);
    const frame = embedPage.frameLocator("iframe");

    await frame.getByRole("row", { name: "Instacart $1.1k" }).click();
    await embedPage.waitForTimeout(500);

    await embedPage.evaluate(() => {
      const iframe = document.querySelector("iframe");
      iframe?.contentWindow?.postMessage({ id: 1337, method: "getState" }, "*");
    });

    await embedPage.waitForTimeout(500);
    expect(
      logMessages.some((msg) =>
        msg.includes(
          `{"id":1337,"result":{"state":"f=advertiser_name+IN+('Instacart')"}}`,
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
          params: "f=advertiser_name+IN+('Instacart')",
        },
        "*",
      );
    });

    await expect(
      frame.getByRole("row", { name: "Instacart $1.1k" }),
    ).toBeVisible();
    expect(
      logMessages.some((msg) => msg.includes(`{"id":1337,"result":true}`)),
    ).toBeTruthy();
  });

  test("navigation works as expected", async ({ embedPage }) => {
    const logMessages: string[] = [];
    await waitForReadyMessage(embedPage, logMessages);
    const frame = embedPage.frameLocator("iframe");

    // Select "Last 6 Hours" as time range
    // Open the menu
    // (Note we cannot use the `interactWithTimeRangeMenu` helper here since its interface is to check the full page)
    await frame.getByLabel("Select time range").click();
    await frame.getByRole("menuitem", { name: "Last 14 Days" }).click();
    // Wait for menu to close
    await expect(
      frame.getByRole("menu", { name: "Select time range" }),
    ).not.toBeVisible();

    // Go to the ` Programmatic Ads Auction ` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitem", { name: "Programmatic Ads Auction", exact: true })
      .click();
    // Time range is still the default
    await expect(frame.getByText("Last 7 Days")).toBeVisible();

    // Go back to the ` Programmatic Ads Bids ` dashboard using the breadcrumbs
    await frame.getByLabel("Breadcrumb dropdown").click();
    await frame
      .getByRole("menuitem", { name: "Programmatic Ads Bids" })
      .click();
    // Old selection has persisted
    await expect(frame.getByText("Last 14 Days")).toBeVisible();

    // Go to `Home` using the breadcrumbs
    await frame.getByText("Home").click();
    // Check that the dashboards are listed
    await expect(
      frame.getByRole("button", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible();
    await expect(
      frame.getByRole("button", { name: "Programmatic Ads Bids" }),
    ).toBeVisible();

    // Go to `Programmatic Ads Auction` using the links on home
    await frame.getByRole("button", { name: "Programmatic Ads Bids" }).click();
    // Old selection has persisted
    await expect(frame.getByText("Last 14 Days")).toBeVisible();
  });
});
