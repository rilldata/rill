import { EXAMPLES } from "@rilldata/web-common/features/welcome/constants";
import { expect } from "playwright/test";
import { test } from "./setup/base";
import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

test.describe("Example project initialization", () => {
  EXAMPLES.forEach((example) => {
    test.describe(`Example project: ${example.title}`, () => {
      test("should initialize new project", async ({ page }) => {
        await page.getByRole("link", { name: example.title }).click();

        const [, fileName] = splitFolderAndFileName(example.firstFile);
        await page.waitForURL(`**/files${example.firstFile}`);

        await expect(
          page.getByRole("heading", { name: fileName }),
        ).toBeVisible();
      });
    });
  });

  test.describe("Empty project", () => {
    test("should initialize new project", async ({ page }) => {
      await page.getByRole("link", { name: "Empty Project" }).click();

      await expect(
        page.getByText("Connect to your data", { exact: true }),
      ).toBeVisible();

      await page.getByRole("link", { name: "rill.yaml" }).click();

      await expect(
        page.getByRole("heading", { name: "rill.yaml" }),
      ).toBeVisible();
    });
  });
});
