import { expect } from "playwright/test";
import { test } from "./setup/base";
import {
  clickMenuButton,
  deleteFile,
  getFileNavEntry,
  openFileNavEntryContextMenu,
} from "./utils/commonHelpers";
import { fileNotPresent, gotoNavEntry } from "./utils/waitHelpers";

test.describe("File Explorer", () => {
  test.use({ project: "Blank" });

  test.describe("File CRUD Operations", () => {
    test.describe.configure({ retries: 2 }); // Problem: flaky tests!
    test("should create, rename, edit, copy, and delete a file", async ({
      page,
    }) => {
      // Create a new file
      await page.getByLabel("Add Asset").click();
      await page.getByRole("menuitem", { name: "More" }).hover();
      await page.getByRole("menuitem", { name: "Blank file" }).click();
      await expect(
        page
          .getByLabel("/untitled_file Nav Entry")
          .getByRole("link", { name: "untitled_file" }),
      ).toBeVisible();
      await expect(
        page.getByRole("heading", { name: "untitled_file", exact: true }),
      ).toBeVisible();

      // Rename the file
      await page.getByRole("listitem", { name: "/untitled_file" }).hover();
      await page.getByLabel("/untitled_file actions menu").click();
      await page.getByRole("menuitem", { name: "Rename" }).click();
      await page.getByLabel("File name").click();
      await page.getByLabel("File name").press("Meta+a");
      await page.getByLabel("File name").fill("README.md");
      await page.getByLabel("File name").press("Enter");
      await expect(
        page.getByRole("link", { name: "README.md" }).nth(1),
      ).toBeVisible();

      // Edit the file
      await page.getByLabel("Auto-save").click(); // Turn off auto-save
      await page.getByRole("textbox").nth(1).click();
      await page.keyboard.type("Here's a README.md file for the e2e test!");
      const saveButton = page.getByRole("button", { name: "Save" });
      await saveButton.click();
      // The Save button disables once the write finishes; wait for that UI
      // feedback so navigating away doesn't race the save.
      await expect(saveButton).toBeDisabled();
      // Navigate away from the file and back to it to verify the changes
      await page.getByRole("link", { name: "rill.yaml" }).click();
      await page.getByRole("link", { name: "README.md" }).first().click();
      await expect(
        page.getByText("Here's a README.md file for the e2e test!"),
      ).toBeVisible();

      // Duplicate the file
      await page.getByRole("listitem", { name: "/README.md" }).hover();
      await page.getByLabel("/README.md actions menu").click();
      await page.getByRole("menuitem", { name: "Duplicate" }).click();
      await expect(
        page.getByRole("link", { name: "README (copy).md" }),
      ).toBeVisible();
      await expect(
        page.getByText("Here's a README.md file for the e2e test!"),
      ).toBeVisible();

      // Delete the file
      await page.getByRole("listitem", { name: "/README.md" }).hover();
      await page.getByLabel("/README.md actions menu").click();
      await page.getByRole("menuitem", { name: "Delete" }).click();
      await page
        .getByRole("alertdialog")
        .getByRole("button", { name: "Delete" })
        .click();
      await expect(
        page.getByRole("link", { name: "README.md" }),
      ).not.toBeVisible();
    });
  });

  test.describe("Folder CRUD Operations", () => {
    test("should create, rename, add sub-folder, and delete the folder", async ({
      page,
    }) => {
      // Create a new folder
      await page.getByLabel("Add Asset").click();
      await page.getByRole("menuitem", { name: "More" }).hover();
      await page.getByRole("menuitem", { name: "Folder" }).click();
      await expect(
        page.getByRole("directory", { name: "untitled_folder" }),
      ).toBeVisible();

      // Rename the folder
      await page.getByRole("directory", { name: "untitled_folder" }).hover();
      await page.getByLabel("untitled_folder actions menu").click();
      await page.getByRole("menuitem", { name: "Rename" }).click();
      await page.getByLabel("Folder name").click();
      await page.getByLabel("Folder name").press("Meta+a");
      await page.getByLabel("Folder name").fill("my-directory");
      await page.getByLabel("Folder name").press("Enter");

      // The rename triggers a reload in the test environment; wait for the
      // renamed entry to render before interacting with it.
      await expect(
        page.getByRole("directory", { name: "my-directory" }),
      ).toBeVisible();

      // Add something to the folder
      await page.getByRole("directory", { name: "my-directory" }).hover();
      await page.getByLabel("my-directory actions menu").click();
      await page.getByRole("menuitem", { name: "New folder" }).hover();
      const [createDirectoryResponse, getFilesResponse] = await Promise.all([
        page.waitForResponse(
          "**/rill.runtime.v1.RuntimeService/CreateDirectory",
        ),
        page.waitForResponse("**/rill.runtime.v1.RuntimeService/ListFiles"),
        page.getByRole("menuitem", { name: "New folder" }).click(),
      ]);

      expect(createDirectoryResponse.status()).toBe(200);
      expect(getFilesResponse.status()).toBe(200);
      const resp = await getFilesResponse.json();
      expect(resp.files.length).toBe(4);
      await expect(
        page.getByRole("directory", {
          name: "my-directory/untitled_folder",
        }),
      ).toBeVisible();

      // Delete the folder
      await page
        .getByRole("button", { name: "my-directory my-directory" })
        .hover();
      await page.getByLabel("my-directory actions menu").click();
      await page.getByRole("menuitem", { name: "Delete" }).click();
      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toContainText(
        "The folder /my-directory and everything in it will be deleted from the project.",
      );
      await dialog.getByRole("button", { name: "Delete" }).click();
      await expect(
        page.getByRole("directory", { name: "my-directory" }),
      ).toHaveCount(0);
    });
  });

  test.describe("Delete confirmation", () => {
    test.use({ project: "AdBids" });
    // The runtime ingests the project before the dashboards resolve,
    // which can take longer than the default 30s locally.
    test.describe.configure({ timeout: 180_000 });

    test("should ask before deleting a canvas or an explore dashboard", async ({
      page,
    }) => {
      const canvasFile = "/dashboards/AdBids_metrics_canvas.yaml";
      const exploreFile = "/dashboards/AdBids_metrics_explore.yaml";
      const dialog = page.getByRole("alertdialog");

      // Open the canvas, so the delete has to leave its page.
      // Its widgets render once the canvas resource has resolved.
      await gotoNavEntry(page, canvasFile);
      await expect(page.getByText("Total records").first()).toBeVisible({
        timeout: 60_000,
      });

      // The dialog names the dashboard and its file, and says whether the delete can be undone
      await openFileNavEntryContextMenu(page, canvasFile);
      await clickMenuButton(page, "Delete");
      await expect(dialog).toContainText("Delete canvas dashboard?");
      await expect(dialog).toContainText(
        `The canvas dashboard AdBids_metrics_canvas and its file ${canvasFile} will be deleted from the project.`,
      );
      await expect(dialog).toContainText("This can't be undone in Rill.");

      // Cancelling keeps the dashboard
      await dialog.getByRole("button", { name: "Cancel" }).click();
      await expect(dialog).toBeHidden();
      await expect(getFileNavEntry(page, canvasFile)).toBeVisible();
      await expect(page).toHaveURL(new RegExp(`/files${canvasFile}`));

      // Confirming deletes it and leaves the deleted file's page
      await deleteFile(page, canvasFile);
      await fileNotPresent(page, canvasFile);
      await expect(page).not.toHaveURL(new RegExp(`/files${canvasFile}`));

      // An explore dashboard gets its own wording
      await openFileNavEntryContextMenu(page, exploreFile);
      await clickMenuButton(page, "Delete");
      await expect(dialog).toContainText("Delete explore dashboard?");
      await expect(dialog).toContainText(
        `The explore dashboard AdBids_metrics_explore and its file ${exploreFile} will be deleted from the project.`,
      );
      await page.keyboard.press("Escape");
      await expect(dialog).toBeHidden();
      await expect(getFileNavEntry(page, exploreFile)).toBeVisible();
    });
  });
});
