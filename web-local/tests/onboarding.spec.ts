import { expect, test } from "@playwright/test";
import path from "node:path";
import { TestDataPath } from "./utils/sourceHelpers";
import { startRuntimeForEachTest } from "./utils/startRuntimeForEachTest";

test.describe("Onboarding", () => {
  startRuntimeForEachTest({ includeRillYaml: false });

  test("Example project", async ({ page }) => {
    await page.goto("/");

    // Click on "Example projects"
    await page.getByRole("link", { name: "Cost monitoring" }).click();

    // Expect to be navigated to the project's first dashboard
    await expect(page).toHaveURL(
      "/files/dashboards/metrics_margin_explore.yaml",
      { timeout: 10_000 },
    );

    // Expect to see the `rill.yaml` file in the sidebar
    await expect(page.getByText("rill.yaml")).toBeVisible();

    // Expect to see the dashboard's default time range
    await expect(page.getByText("Last 3 weeks")).toBeVisible({
      timeout: 15_000,
    });
  });

  test("Rill-managed OLAP - local file", async ({ page }) => {
    test.setTimeout(20_000);
    await page.goto("/");

    // Should be redirected to the onboarding page
    await expect(page).toHaveURL("/welcome");
    await expect(page.getByText("Welcome to Rill")).toBeVisible();

    // Click on "Connect your data"
    await page.getByRole("button", { name: "Connect your data" }).click();

    // Should get redirected to the select connectors page
    await expect(page).toHaveURL("/welcome/select-connectors");

    // Click on button with aria-label "Local file"
    await page.getByLabel("local_file").click();

    // Click on "Continue"
    await page.getByRole("button", { name: "Continue" }).click();

    // Should get redirected to the add credentials page
    await expect(page).toHaveURL("/welcome/add-credentials");

    // Upload a file (Note: this is copied from `sourceHelpers.ts` and should be cleaned-up.)
    const [fileChooser] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByText("Upload a CSV, JSON or Parquet file").click(),
    ]);
    const fileUploadPromise = fileChooser.setFiles([
      path.join(TestDataPath, "Adbids.csv"),
    ]);
    const fileRespWaitPromise = page.waitForResponse(/files\/entry/);
    await Promise.all([fileUploadPromise, fileRespWaitPromise]);

    // Expect to be navigated to the new source page
    await expect(page).toHaveURL("/files/sources/Adbids.yaml");

    // Click "view this source"
    await page.getByRole("button", { name: "View this source" }).click();

    // Expect to see data
    await expect(page.getByLabel("Results Preview Table")).toBeVisible();
    await expect(
      page.getByLabel("Results Preview Table").getByText("timestamp"), // timestamp is the name of the second column
    ).toBeVisible();
    await expect(
      page.getByLabel("Results Preview Table").getByText("4000", {
        exact: true,
      }), // 4000 is the first row's ID
    ).toBeVisible();
    await expect(
      page.getByLabel("Results Preview Table").getByText("Facebook").first(), // Facebook is one of the publishers
    ).toBeVisible();
  });

  // TODO: Spin-up a local ClickHouse instance, so we can test this actually working.
  test("Self-managed OLAP - ClickHouse", async ({ page }) => {
    await page.goto("/");

    // Should be redirected to the onboarding page
    await expect(page).toHaveURL("/welcome");
    await expect(page.getByText("Welcome to Rill")).toBeVisible();

    // Pick a self-managed ClickHouse OLAP
    await page.getByRole("button", { name: "Connect your data" }).click();
    await page.getByRole("button", { name: "Self-managed" }).click();
    // "ClickHouse" is the default, so no need to pick it explicitly.
    await page.getByRole("button", { name: "Continue" }).click();

    // Add credentials
    await page.getByRole("textbox", { name: "Host" }).click();
    await page.getByRole("textbox", { name: "Host" }).fill("localhost");
    await page.getByRole("textbox", { name: "Host" }).press("Tab");
    await page.getByRole("textbox", { name: "Port (optional)" }).fill("9000");

    // Submit form and expect to see an error connecting to the dummy ClickHouse instance
    await page.getByRole("button", { name: "Connect" }).click();
    await expect(
      page
        .locator("#add-data-form div")
        .filter({ hasText: "dial tcp [::1]:9000: connect" })
        .nth(1),
    ).toBeVisible();
  });
});
