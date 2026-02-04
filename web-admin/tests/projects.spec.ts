import { expect } from "@playwright/test";
import { test } from "./setup/base";
import {
  RILL_ORG_NAME,
  RILL_PROJECT_NAME,
  RILL_PROJECT_DISPLAY_NAME,
} from "./setup/constants";

test.describe("Projects", () => {
  test("admins should see the admin-only pages", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb");
    await expect(adminPage.getByRole("link", { name: "Status" })).toBeVisible();
    await expect(
      adminPage.getByRole("link", { name: "Settings" }),
    ).toBeVisible();
  });

  test.describe("Settings", () => {
    const settingsUrl = `/${RILL_ORG_NAME}/${RILL_PROJECT_NAME}/-/settings`;

    test("should display the settings page with all sections", async ({
      adminPage,
    }) => {
      await adminPage.goto(settingsUrl);

      // Check that the page title sections are visible
      await expect(
        adminPage.getByText("Project", { exact: true }),
      ).toBeVisible();
      await expect(adminPage.getByText("Danger zone")).toBeVisible();

      // Check that all settings sections are visible
      await expect(adminPage.getByText("Public visibility")).toBeVisible();
      await expect(adminPage.getByText("Hibernate project")).toBeVisible();
      await expect(adminPage.getByText("Delete project")).toBeVisible();
    });

    test("should display current project name", async ({ adminPage }) => {
      await adminPage.goto(settingsUrl);

      const nameInput = adminPage.locator("#name");
      await expect(nameInput).toHaveValue(RILL_PROJECT_DISPLAY_NAME);
    });

    test("should show warning when renaming project", async ({ adminPage }) => {
      await adminPage.goto(settingsUrl);

      const nameInput = adminPage.locator("#name");
      await nameInput.fill("Test Project Rename");

      await expect(
        adminPage.getByText(
          "Renaming this project will invalidate all existing URLs and shared links.",
        ),
      ).toBeVisible();
    });

    test("should enable Save button when changes are made", async ({
      adminPage,
    }) => {
      await adminPage.goto(settingsUrl);

      const saveButton = adminPage.getByRole("button", { name: "Save" });
      await expect(saveButton).toBeDisabled();

      const descriptionInput = adminPage.locator("#description");
      await descriptionInput.fill("Test description change");

      await expect(saveButton).toBeEnabled();
    });

    test("should update project description successfully", async ({
      adminPage,
    }) => {
      await adminPage.goto(settingsUrl);

      const descriptionInput = adminPage.locator("#description");
      const saveButton = adminPage.getByRole("button", { name: "Save" });

      const originalDescription = await descriptionInput.inputValue();

      const testDescription = `E2E test description - ${Date.now()}`;
      await descriptionInput.fill(testDescription);
      await saveButton.click();

      await expect(adminPage.getByLabel("Notification")).toHaveText(
        "Updated project",
      );

      // Restore the original description
      await descriptionInput.fill(originalDescription);
      await saveButton.click();

      await expect(adminPage.getByLabel("Notification")).toHaveText(
        "Updated project",
      );
    });

    test("should display current visibility status", async ({ adminPage }) => {
      await adminPage.goto(settingsUrl);

      const isPublicText = adminPage.getByText("This project is currently", {
        exact: false,
      });
      await expect(isPublicText).toBeVisible();
    });

    test("should show confirmation dialog when making project public", async ({
      adminPage,
    }) => {
      await adminPage.goto(settingsUrl);

      const makePublicButton = adminPage.getByRole("button", {
        name: "Make public",
      });

      // Only run if project is private
      if (await makePublicButton.isVisible()) {
        await makePublicButton.click();

        await expect(
          adminPage.getByText("Make this project public?"),
        ).toBeVisible();

        await adminPage.getByRole("button", { name: "Cancel" }).click();
      }
    });

    test("should show hibernate confirmation dialog", async ({ adminPage }) => {
      await adminPage.goto(settingsUrl);

      await adminPage
        .getByRole("button", { name: "Hibernate project" })
        .click();

      await expect(
        adminPage.getByText("Hibernate this project?"),
      ).toBeVisible();

      await expect(
        adminPage.getByText(`Type hibernate ${RILL_PROJECT_NAME}`),
      ).toBeVisible();

      await expect(
        adminPage.getByRole("button", { name: "Continue" }),
      ).toBeDisabled();

      // Enter confirmation text
      const confirmInput = adminPage.locator("#confirmation");
      await confirmInput.fill(`hibernate ${RILL_PROJECT_NAME}`);

      await expect(
        adminPage.getByRole("button", { name: "Continue" }),
      ).toBeEnabled();

      await adminPage.getByRole("button", { name: "Cancel" }).click();
    });

    test("should show delete confirmation dialog", async ({ adminPage }) => {
      await adminPage.goto(settingsUrl);

      await adminPage.getByRole("button", { name: "Delete project" }).click();

      await expect(adminPage.getByText("Delete this project?")).toBeVisible();

      await expect(
        adminPage.getByText(`Type delete ${RILL_PROJECT_NAME}`),
      ).toBeVisible();

      await expect(
        adminPage.getByRole("button", { name: "Continue" }),
      ).toBeDisabled();

      // Enter confirmation text
      const confirmInput = adminPage.locator("#confirmation");
      await confirmInput.fill(`delete ${RILL_PROJECT_NAME}`);

      await expect(
        adminPage.getByRole("button", { name: "Continue" }),
      ).toBeEnabled();

      await adminPage.getByRole("button", { name: "Cancel" }).click();
    });
  });
});
