import { expect } from "@playwright/test";
import {
  extractOpenLink,
  getCompleteEmail,
  getOpenLinkFromEmail,
  waitForEmail,
} from "@rilldata/web-common/tests/utils/email-utils.ts";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions.ts";
import { EMBED_HOST_ORIGIN } from "./embed-host/constants.ts";
import type { EmbedHostApp } from "./embed-host/embed-host-app.ts";
import { getEmbedUser, getExtraRecipientEmail } from "./embed-host/users.ts";
import { test } from "./setup/base";

// Unique per run, so reports left behind by a failed run don't break the next one.
const RUN_ID = Date.now();
const ALICE_REPORT = `Alice embed report ${RUN_ID}`;
const CAROL_REPORT = `Carol embed report ${RUN_ID}`;

// Shared across the serial tests below.
let aliceOpenLink = "";

test.describe.serial("Embedded reports", () => {
  test("Should not offer reports without an external user", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "none" });
    await openExportMenu(embedHost);

    await expect(
      embedHost.frame.getByRole("menuitem", { name: "Export as CSV" }),
    ).toBeVisible();
    await expect(
      embedHost.frame.getByRole("menuitem", {
        name: "Create scheduled report...",
      }),
    ).not.toBeVisible();
  });

  test("Should create a report", async ({ embedHost }) => {
    test.setTimeout(60_000);
    const alice = getEmbedUser("alice");

    await embedHost.goto({ user: "alice" });
    await openReportForm(embedHost);

    const reportForm = embedHost.frame.locator("form#scheduled-report-form");
    await reportForm.getByTitle("Report title").fill(ALICE_REPORT);

    // Set as a daily report at 10:00 PM
    await reportForm.getByLabel("Frequency").click();
    await embedHost.frame.getByRole("option", { name: "Daily" }).click();
    await reportForm.getByLabel("Time", { exact: true }).click();
    await embedHost.frame.getByRole("option", { name: "10:00 PM" }).click();

    await interactWithTimeRangeMenu(reportForm, async () => {
      // Menu content is portaled to the iframe's body
      await embedHost.frame
        .getByRole("menuitem", { name: "Last 14 days" })
        .click();
    });

    // Add "Domain" as a row; the dimension table's "Publisher" is a column
    await reportForm.getByLabel("Add Rows fields").click();
    await embedHost.frame.getByRole("menuitem", { name: "Domain" }).click();
    await expect(reportForm.getByLabel("Rows field list")).toHaveText(/Domain/);
    await expect(reportForm.getByLabel("Columns field list")).toHaveText(
      /Publisher\s*Total records\s*Sum of Bid Price/,
    );

    // Alice and an extra recipient get the report
    await addEmailRecipient(embedHost, alice.email);
    await addEmailRecipient(embedHost, getExtraRecipientEmail());

    // Alice's email isn't known to Rill, so she enters it to get the open link
    const ownerEmail = reportForm.getByLabel("Your email");
    await expect(ownerEmail).toBeEnabled();
    await ownerEmail.fill(alice.email);

    await embedHost.frame.getByLabel("Create report").scrollIntoViewIfNeeded();
    await embedHost.frame.getByLabel("Create report").click();
    await expect(embedHost.frame.getByLabel("Notification")).toContainText(
      "Report created",
    );

    await embedHost.navigate("reports");
    await expect(
      embedHost.frame.getByRole("link", { name: ALICE_REPORT }),
    ).toBeVisible();
  });

  test("Should only show the report to its creator and admins", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "bob" });
    await embedHost.navigate("reports");
    await expect(
      embedHost.frame.getByText("You don't have any reports yet"),
    ).toBeVisible();
    await expect(
      embedHost.frame.getByRole("link", { name: ALICE_REPORT }),
    ).not.toBeVisible();

    await embedHost.goto({ user: "admin" });
    await embedHost.navigate("reports");
    await expect(
      embedHost.frame.getByRole("link", { name: ALICE_REPORT }),
    ).toBeVisible();
  });

  test("Should run the report and send emails", async ({ embedHost }) => {
    test.setTimeout(60_000);
    const alice = getEmbedUser("alice");
    const extraRecipient = getExtraRecipientEmail();

    const time = await runReport(embedHost, "alice", ALICE_REPORT);

    // Only the creator's email has an open link, and it points to the embed host's open page
    const aliceEmail = await getCompleteEmail(
      await waitForEmail(ALICE_REPORT, time, alice.email),
    );
    aliceOpenLink = extractOpenLink(aliceEmail) ?? "";
    expect(aliceOpenLink).toMatch(
      new RegExp(`^${EMBED_HOST_ORIGIN}/open\\.html\\?rill_open=`),
    );

    const extraEmail = await getCompleteEmail(
      await waitForEmail(ALICE_REPORT, time, extraRecipient),
    );
    expect(extractOpenLink(extraEmail)).toBeUndefined();

    // Both have the export attached, but only the extra recipient has an unsubscribe link,
    // since the creator can't unsubscribe from their own report
    for (const email of [aliceEmail, extraEmail]) {
      expect(email.Attachments.length).toBeGreaterThan(0);
    }
    expect(aliceEmail.Text).not.toContain("To unsubscribe from this report");
    expect(extraEmail.Text).toContain("To unsubscribe from this report");
  });

  test("Should open the report as its creator", async ({ embedHost }) => {
    await embedHost.openLink(aliceOpenLink, "alice");

    // Lands on the explore's pivot with the report's rows and columns
    await expect(embedHost.frame.getByLabel("Drag list rows")).toHaveText(
      /Domain/,
    );
    await expect(embedHost.frame.getByLabel("Drag list columns")).toHaveText(
      /Publisher\s*Total records\s*Sum of Bid Price/,
    );
    await expect(
      embedHost.frame.getByRole("link", { name: "Home" }),
    ).toBeVisible();
  });

  test("Should not open the report as another user", async ({ embedHost }) => {
    await embedHost.openLink(aliceOpenLink, "bob");

    await embedHost.expectError(
      "was not created by the provided external_user_id",
    );
    await expect(embedHost.page.locator("#rill-frame")).toBeHidden();
  });

  test("Should open the report with navigation disabled", async ({
    embedHost,
  }) => {
    await embedHost.openLink(aliceOpenLink, "alice");
    await embedHost.setNavigation(false);

    await expect(embedHost.frame.getByLabel("Drag list rows")).toHaveText(
      /Domain/,
    );
    // Without navigation, the embed has no navigation bar to leave the report's dashboard
    await expect(
      embedHost.frame.getByRole("link", { name: "Home" }),
    ).not.toBeVisible();
  });

  test("Should lock the creator email passed as user_email", async ({
    embedHost,
  }) => {
    test.setTimeout(60_000);
    const carol = getEmbedUser("carol");

    await embedHost.goto({ user: "carol" });
    await openReportForm(embedHost);

    const reportForm = embedHost.frame.locator("form#scheduled-report-form");
    await reportForm.getByTitle("Report title").fill(CAROL_REPORT);

    const ownerEmail = reportForm.getByLabel("Your email");
    await expect(ownerEmail).toHaveValue(carol.email);
    await expect(ownerEmail).toBeDisabled();

    await embedHost.frame.getByLabel("Create report").scrollIntoViewIfNeeded();
    await embedHost.frame.getByLabel("Create report").click();
    await expect(embedHost.frame.getByLabel("Notification")).toContainText(
      "Report created",
    );

    const time = await runReport(embedHost, "carol", CAROL_REPORT);
    const link = await getOpenLinkFromEmail(
      await waitForEmail(CAROL_REPORT, time, carol.email),
    );
    expect(link).toMatch(
      new RegExp(`^${EMBED_HOST_ORIGIN}/open\\.html\\?rill_open=`),
    );
  });

  test("Should edit the report", async ({ embedHost }) => {
    const alice = getEmbedUser("alice");

    await gotoReport(embedHost, "alice", ALICE_REPORT);
    await embedHost.frame.getByLabel("Report context menu").click();
    await embedHost.frame
      .getByRole("menuitem", { name: "Edit Report" })
      .click();

    const reportForm = embedHost.frame.locator("form#scheduled-report-form");
    // The saved creator email is kept
    await expect(reportForm.getByLabel("Your email")).toHaveValue(alice.email);

    await reportForm.getByLabel("Frequency").click();
    await embedHost.frame.getByRole("option", { name: "Monthly" }).click();
    await embedHost.frame.getByLabel("Save report").click();

    await expect(embedHost.frame.getByLabel("Notification")).toHaveText(
      "Report edited",
    );
    await expect(embedHost.frame.getByLabel("Report schedule")).toHaveText(
      /Repeats\s+At 10:00 PM, on the 1st of each month/m,
    );
  });

  test("Should delete the reports", async ({ embedHost }) => {
    for (const [user, title] of [
      ["alice", ALICE_REPORT],
      ["carol", CAROL_REPORT],
    ] as const) {
      await gotoReport(embedHost, user, title);
      await embedHost.frame.getByLabel("Report context menu").click();
      await embedHost.frame
        .getByRole("menuitem", { name: "Delete Report" })
        .click();

      await expect(
        embedHost.frame.getByText("You don't have any reports yet"),
      ).toBeVisible();
    }
  });
});

async function openExportMenu(embedHost: EmbedHostApp) {
  // Enter the "Publisher" dimension table
  await embedHost.frame
    .getByLabel("Open dimension details")
    .filter({ hasText: "Publisher" })
    .click();
  await embedHost.frame.getByLabel("Dimension Display").waitFor();

  // Clicking the dimension can show a tooltip over the export button; moving the mouse away dismisses it.
  await embedHost.page.mouse.move(0, 0);

  await embedHost.frame.getByLabel("Export dimension table data").click();
}

async function openReportForm(embedHost: EmbedHostApp) {
  await openExportMenu(embedHost);
  await embedHost.frame
    .getByRole("menuitem", { name: "Create scheduled report..." })
    .click();
}

async function addEmailRecipient(embedHost: EmbedHostApp, email: string) {
  // The recipients input's label isn't associated with its last, empty input.
  const input = embedHost.frame
    .locator("form#scheduled-report-form")
    .locator('input[id^="emailRecipients."]')
    .last();
  await input.fill(email);
  await input.press("Enter");
}

async function gotoReport(
  embedHost: EmbedHostApp,
  user: "alice" | "carol",
  title: string,
) {
  await embedHost.goto({ user });
  await embedHost.navigate("reports");
  await embedHost.frame.getByRole("link", { name: title }).click();
  await expect(embedHost.frame.getByLabel("Report name")).toHaveText(title);
}

/** Runs the report now, and returns the time before the run, to find its emails. */
async function runReport(
  embedHost: EmbedHostApp,
  user: "alice" | "carol",
  title: string,
) {
  await gotoReport(embedHost, user, title);
  const time = new Date();
  await embedHost.frame.getByRole("button", { name: "Run now" }).click();
  await expect(embedHost.frame.getByLabel("Notification")).toHaveText(
    "Triggered an ad-hoc run of this report.",
  );
  return time;
}
