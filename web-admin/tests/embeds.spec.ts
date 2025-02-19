import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Embeds", () => {
  test("embeds should load", async ({ embedPage }) => {
    const frame = embedPage.frameLocator("iframe");
    // Set the time zone to UTC
    await frame.getByLabel("Timezone selector").click();
    await frame.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();

    // Check the Big Number
    await expect(
      frame.getByRole("button", { name: "Advertising Spend Overall $1.30M" }),
    ).toBeVisible();
  });

  test("state is emitted for embeds", async ({ embedPage }) => {
    const logMessages: string[] = [];

    const waitForReadyMessage = new Promise<void>((resolve) => {
      embedPage.on("console", async (msg) => {
        if (msg.type() === "log") {
          const args = await Promise.all(msg.args().map((arg) => arg.jsonValue()));
          const logMessage = JSON.stringify(args);
          logMessages.push(logMessage);

          if (logMessage.includes(`{"method":"ready"}`)) {
            resolve();
          }
        }
      });
    });

    await waitForReadyMessage;

    const frame = embedPage.frameLocator("iframe");

    // Set the time zone to UTC
    await frame.getByLabel("Timezone selector").click();
    await frame.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();

    // Click a dimension row
    await frame.getByRole("row", { name: "Instacart $107.3k" }).click();
    await embedPage.waitForTimeout(500);

    expect(
      logMessages.some((msg) =>
        msg.includes("tz=UTC&f=advertiser_name+IN+('Instacart')"),
      ),
    ).toBeTruthy();
  });

  test("getState returns from embed", async ({ embedPage }) => {
    const logMessages: string[] = [];

    const waitForReadyMessage = new Promise<void>((resolve) => {
      embedPage.on("console", async (msg) => {
        if (msg.type() === "log") {
          const args = await Promise.all(msg.args().map((arg) => arg.jsonValue()));
          const logMessage = JSON.stringify(args);
          logMessages.push(logMessage);

          if (logMessage.includes(`{"method":"ready"}`)) {
            resolve();
          }
        }
      });
    });

    await waitForReadyMessage;

    const frame = embedPage.frameLocator("iframe");

    // Set the time zone to UTC
    await frame.getByLabel("Timezone selector").click();
    await frame.getByRole("menuitem", { name: "UTC GMT +00:00 UTC" }).click();

    // Click a dimension row
    await frame.getByRole("row", { name: "Instacart $107.3k" }).click();
    await embedPage.waitForTimeout(500);

    await embedPage.evaluate(() => {
      const iframe = document.querySelector("iframe");
      if (iframe) {
        if (iframe.contentWindow) {
          iframe.contentWindow.postMessage(
            { id: 1337, method: "getState" },
            "*",
          );
        }
      }
    });
    await embedPage.waitForTimeout(500);
    expect(
      logMessages.some((msg) =>
        msg.includes(
          `{"id":1337,"result":{"state":"tz=UTC&f=advertiser_name+IN+('Instacart')"}}`,
        ),
      ),
    ).toBeTruthy();
  });

  test("setState changes embedded explore", async ({ embedPage }) => {
    const logMessages: string[] = [];

    const waitForReadyMessage = new Promise<void>((resolve) => {
      embedPage.on("console", async (msg) => {
        if (msg.type() === "log") {
          const args = await Promise.all(msg.args().map((arg) => arg.jsonValue()));
          const logMessage = JSON.stringify(args);
          logMessages.push(logMessage);

          if (logMessage.includes(`{"method":"ready"}`)) {
            resolve(); // ✅ Resolve promise when "ready" appears
          }
        }
      });
    });

    await waitForReadyMessage;

    const frame = embedPage.frameLocator("iframe");

    await embedPage.evaluate(() => {
      const iframe =
        window.document.getElementsByTagName("iframe")[0].contentWindow;
      if (iframe) {
        iframe.postMessage(
          {
            id: 1337,
            method: "setState",
            params: "tz=UTC&f=advertiser_name+IN+('Instacart')",
          },
          "*",
        );
      }
    });

    await expect(frame.getByLabel("Timezone selector")).toHaveText("UTC");
    await expect(frame.getByRole("row", { name: "Instacart $107.3k" })).toBeVisible();
    expect(logMessages.some((msg) => msg.includes(`{"id":1337,"result":true}`))).toBeTruthy();
  });


});
