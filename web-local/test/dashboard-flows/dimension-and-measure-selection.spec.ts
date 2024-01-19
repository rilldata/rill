import { expect, test } from "@playwright/test";
import { useDashboardFlowTestSetup } from "web-local/test/dashboard-flows/dashboard-flow-test-setup";
import { startRuntimeForEachTest } from "../utils/startRuntimeForEachTest";

test.describe("dimension and measure selectors", () => {
  startRuntimeForEachTest();
  // dashboard test setup
  useDashboardFlowTestSetup();

  test("dimension and measure selectors flow", async ({ page }) => {
    await page.getByRole("button", { name: "All Measures" }).click();
    await page.getByRole("menuitem", { name: "Total records" }).click();
    await page.getByRole("button", { name: "1 of 2 Measures" }).click();

    await expect(page.getByText("Sum(bid_price) 300.6k")).toBeVisible();
    await expect(page.getByText("Total records 100.0k")).not.toBeVisible();

    await page.getByRole("button", { name: "1 of 2 Measures" }).click();
    await page.getByRole("menuitem", { name: "Total records" }).click();
    await page.getByRole("menuitem", { name: "Sum(bid_price)" }).click();
    await page.getByRole("button", { name: "1 of 2 Measures" }).click();

    await expect(page.getByText("Sum(bid_price) 300.6k")).not.toBeVisible();
    await expect(page.getByText("Total records 100.0k")).toBeVisible();

    await page.getByRole("button", { name: "All Dimensions" }).click();
    await page.getByRole("menuitem", { name: "Publisher" }).click();
    await page.getByRole("button", { name: "1 of 2 Dimensions" }).click();

    await expect(page.getByText("Publisher")).not.toBeVisible();
    await expect(page.getByText("Domain")).toBeVisible();

    await page.getByRole("button", { name: "1 of 2 Dimensions" }).click();
    await page.getByRole("menuitem", { name: "Publisher" }).click();
    await page.getByRole("menuitem", { name: "Domain" }).click();
    await page.getByRole("button", { name: "1 of 2 Dimensions" }).click();

    await expect(page.getByText("Publisher")).toBeVisible();
    await expect(page.getByText("Domain")).not.toBeVisible();
  });
});
