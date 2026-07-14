import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("custom viz components", () => {
  test.use({ project: "AdBidsComponents" });

  test("can create a blank custom viz with a workspace and test bindings", async ({
    page,
  }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();

    // "Custom viz" opens the examples gallery directly; "Start blank" sits in
    // its header for the from-scratch path.
    await page.getByRole("menuitem", { name: "Custom viz" }).click();
    await page.getByRole("button", { name: "Start blank" }).click();

    // Lands on the new component file in the component workspace.
    await expect(page).toHaveURL(/viz_library\/component\.yaml/);

    // The inspector shows the declared params as test bindings.
    await expect(page.getByText("Test values")).toBeVisible();
    await expect(page.getByText("Metrics View", { exact: true })).toBeVisible();
    await expect(page.getByText("Measure", { exact: true })).toBeVisible();
    await expect(page.getByText("Dimension", { exact: true })).toBeVisible();

    // The pre-existing component is not used by any dashboards yet.
    await expect(page.getByText("Used by", { exact: true })).toBeVisible();

    // The workspace opens in the visual view by default; the preview renders a
    // chart against the auto-selected metrics view and test bindings.
    await expect(page.locator(".vega-embed svg, .vega-embed canvas")).toBeVisible(
      { timeout: 15_000 },
    );

    // Switching to the code view shows the starter YAML.
    await page.getByLabel("Switch to code editor").click();
    await expect(page.getByText("type: component")).toBeVisible();
  });

  test("can add a custom viz to a canvas with generated param inputs", async ({
    page,
  }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();
    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();

    // The project's standalone components are listed in their own section.
    await page.getByRole("menuitem", { name: "Your components" }).hover();
    await page.getByRole("menuitem", { name: "Measure trend" }).click();

    // The inspector shows the generated param form with smart defaults bound.
    await expect(page.getByText("Metrics View", { exact: true })).toBeVisible();
    await expect(page.getByText("Measure", { exact: true })).toBeVisible();
    await expect(page.getByText("Time Dim", { exact: true })).toBeVisible();

    // The referenced component's header links to editing the component file.
    await expect(page.getByText("Edit component")).toBeVisible();
    await expect(
      page.getByText("Edits affect all dashboards using this component."),
    ).toBeVisible();

    // The referenced chart renders.
    await expect(
      page.locator(".vega-embed svg, .vega-embed canvas").first(),
    ).toBeVisible({ timeout: 15_000 });
  });

  test("can reference the same component twice with different bindings", async ({
    page,
  }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();
    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();

    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();
    await page.getByRole("menuitem", { name: "Your components" }).hover();
    await page.getByRole("menuitem", { name: "Measure trend" }).click();

    // Wait for the first instance to render before inserting the second.
    await expect(page.locator(".vega-embed").first()).toBeVisible({
      timeout: 15_000,
    });

    await page
      .getByRole("button", { name: "Resize row 1 column 1" })
      .hover({ force: true });
    await page
      .getByRole("button", {
        name: "Insert widget in row 1 at column 2",
        exact: true,
      })
      .click();
    await page.getByRole("menuitem", { name: "Your components" }).click();
    await page
      .getByRole("menuitem", { name: "Measure trend" })
      .click({ timeout: 10_000 });

    // Both instances render independently.
    await expect(page.locator(".vega-embed")).toHaveCount(2, {
      timeout: 15_000,
    });
  });
});

test.describe("custom viz flag off", () => {
  test.use({ project: "AdBids" });

  test("menus do not offer custom viz when the flag is off", async ({
    page,
  }) => {
    await page.getByLabel("Add asset").waitFor({ state: "visible" });
    await page.getByLabel("Add asset").click();
    await expect(
      page.getByRole("menuitem", { name: "Custom viz" }),
    ).toHaveCount(0);
    await page.keyboard.press("Escape");

    await page.getByLabel("Add asset").click();
    await page.getByRole("menuitem", { name: "Canvas dashboard" }).click();
    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();
    await expect(
      page.getByRole("menuitem", { name: "Your components" }),
    ).toHaveCount(0);
  });
});
