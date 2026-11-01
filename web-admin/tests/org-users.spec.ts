import { expect, type Page } from "@playwright/test";
import { execAsync } from "@rilldata/web-common/tests/utils/spawn";
import dotenv from "dotenv";
import path from "path";
import { fileURLToPath } from "url";
import { test } from "./setup/base";
import { RILL_ORG_NAME } from "./setup/constants";

// The users page loads invites 50 at a time. Invite more than one page so the
// search and filter tests have to reach rows that are not loaded initially.
const INVITE_PAGE_SIZE = 50;
const INVITE_COUNT = INVITE_PAGE_SIZE + 10;

test.describe.serial("Org users search and filters", () => {
  // Load environment variables from our root `.env` file
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  dotenv.config({ path: path.resolve(__dirname, "../../.env") });

  const adminEmail = process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL ?? "";
  const [adminLocalPart, adminDomain] = adminEmail.split("@");

  // Plus-addressed variants of the admin account, e.g. `e2e-admin+invite-007@rilldata.com`.
  // These accounts have not signed up, so they show up as pending invites.
  const inviteEmail = (suffix: string) =>
    `${adminLocalPart}+invite-${suffix}@${adminDomain}`;
  // Invites are listed in email order, so `zz` sorts this one after every
  // numbered invite and it only exists beyond the first page.
  const targetEmail = inviteEmail("zz-target");
  const inviteEmails = [
    ...Array.from({ length: INVITE_COUNT - 1 }, (_, i) =>
      inviteEmail(String(i).padStart(3, "0")),
    ),
    targetEmail,
  ];

  const usersUrl = `/${RILL_ORG_NAME}/-/users`;
  const pendingRows = (page: Page) =>
    page.getByRole("row", { name: "Pending invitation" });
  const searchBox = (page: Page) =>
    page.getByRole("textbox", { name: "Search" });

  async function selectFilter(page: Page, trigger: string, option: string) {
    await page.getByRole("button", { name: trigger, exact: true }).click();
    await page.getByRole("menuitemcheckbox", { name: option }).click();
  }

  test.beforeAll(async () => {
    test.setTimeout(120_000);
    if (!adminEmail) {
      throw new Error(
        "Missing required environment variables for authentication",
      );
    }

    await execAsync(
      `rill sudo quota set --org ${RILL_ORG_NAME} --outstanding-invites ${INVITE_COUNT * 2}`,
    );
    // Invite in small parallel batches: fast enough without flooding the admin service.
    for (let i = 0; i < inviteEmails.length; i += 10) {
      await Promise.all(
        inviteEmails
          .slice(i, i + 10)
          .map((email) =>
            execAsync(
              `rill user add --org ${RILL_ORG_NAME} --email ${email} --role viewer`,
            ),
          ),
      );
    }
  });

  test.afterAll(async () => {
    test.setTimeout(120_000);
    // Teardown deletes the org, but remove the invites anyway so the suite can
    // be re-run against a long-lived dev environment.
    for (let i = 0; i < inviteEmails.length; i += 10) {
      await Promise.all(
        inviteEmails
          .slice(i, i + 10)
          .map((email) =>
            execAsync(
              `rill user remove --org ${RILL_ORG_NAME} --email ${email}`,
            ).catch(() => undefined),
          ),
      );
    }
  });

  test("should find a pending invite beyond the first page by search", async ({
    page,
  }) => {
    await page.goto(usersUrl);

    // Only the first page of invites is loaded initially, without the target.
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();
    await expect(pendingRows(page)).toHaveCount(INVITE_PAGE_SIZE);
    await expect(page.getByRole("row", { name: targetEmail })).toHaveCount(0);

    await searchBox(page).fill("zz-target");

    await expect(page.getByRole("row", { name: targetEmail })).toBeVisible();
    await expect(pendingRows(page)).toHaveCount(1);
    await expect(page.getByRole("row", { name: adminEmail })).toHaveCount(0);

    // Clearing the search restores the unfiltered list. The invite pages
    // fetched for the search stay cached, so every invite is listed now.
    await searchBox(page).fill("");
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();
    await expect(pendingRows(page)).toHaveCount(INVITE_COUNT);
  });

  test("should show every pending invite under the pending invites filter", async ({
    page,
  }) => {
    await page.goto(usersUrl);
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();

    await selectFilter(page, "All users", "Pending invites");

    // Every page of invites is loaded, not just the first one.
    await expect(pendingRows(page)).toHaveCount(INVITE_COUNT);
    await expect(page.getByRole("row", { name: targetEmail })).toBeVisible();
    await expect(page.getByRole("row", { name: adminEmail })).toHaveCount(0);

    await selectFilter(page, "Pending invites", "Members");
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();
    await expect(pendingRows(page)).toHaveCount(0);
  });

  test("should combine search with the role filter", async ({ page }) => {
    await page.goto(usersUrl);
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();

    // Matches the ten invites numbered 000 to 009 and no members.
    await searchBox(page).fill("invite-00");
    await expect(pendingRows(page)).toHaveCount(10);
    await expect(page.getByRole("row", { name: adminEmail })).toHaveCount(0);

    // The invites were created as viewers, so the viewer role keeps them.
    await selectFilter(page, "All Roles", "Viewers");
    await expect(pendingRows(page)).toHaveCount(10);

    // No admin matches the search, so the table shows the empty state.
    await selectFilter(page, "Viewers", "Admins");
    await expect(page.getByText("No users found")).toBeVisible();

    // Clearing the search with the admin role still selected shows the admin.
    await searchBox(page).fill("");
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();
    await expect(pendingRows(page)).toHaveCount(0);

    // Searching for the admin by email keeps the row.
    await searchBox(page).fill(adminEmail);
    await expect(page.getByRole("row", { name: adminEmail })).toBeVisible();
  });
});
