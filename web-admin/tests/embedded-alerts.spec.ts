import { expect } from "@playwright/test";
import {
  extractOpenLink,
  getCompleteEmail,
  waitForEmail,
} from "@rilldata/web-common/tests/utils/email-utils.ts";
import { getEmbedHostConfig } from "./embed-host/config.ts";
import type { EmbedHostApp } from "./embed-host/embed-host-app.ts";
import { getEmbedUser, getExtraRecipientEmail } from "./embed-host/users.ts";
import { test } from "./setup/base";

// Unique per run, so alerts left behind by a failed run don't break the next one.
const RUN_ID = Date.now();
const ALICE_ALERT = `Alice embed alert ${RUN_ID}`;
const ALICE_ALERT_EDITED = `Alice edited embed alert ${RUN_ID}`;

// Shared across the serial tests below.
let aliceOpenLink = "";

test.describe.serial("Embedded alerts", () => {
  test("Should not offer alerts without an external user", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "none" });
    await expect(
      embedHost.frame.getByLabel("Open dimension details").first(),
    ).toBeVisible();
    await expect(
      embedHost.frame.getByRole("button", { name: "Alerts", exact: true }),
    ).not.toBeVisible();
  });

  test("Should create an alert", async ({ embedHost }) => {
    test.setTimeout(60_000);
    const alice = getEmbedUser("alice");

    await embedHost.goto({ user: "alice" });
    await openAlertsMenu(embedHost);
    await embedHost.frame
      .getByRole("menuitem", { name: "Create alert" })
      .click();

    const alertForm = embedHost.frame.locator("form#create-alert-form");
    // Keep the default measure and criteria (greater than 0), so the alert triggers right away
    await alertForm.getByRole("button", { name: "Next" }).click();
    await alertForm.getByRole("button", { name: "Next" }).click();
    await alertForm.getByTitle("Alert name").fill(ALICE_ALERT);

    // An extra recipient gets the alert
    const recipients = alertForm.locator('input[id^="emailRecipients."]');
    await recipients.last().fill(getExtraRecipientEmail());
    await recipients.last().press("Enter");

    // Alice's email isn't known to Rill, so she enters it to get the open link
    const ownerEmail = alertForm.getByLabel("Your email");
    await expect(ownerEmail).toBeEnabled();
    await ownerEmail.fill(alice.email);

    await alertForm.getByRole("button", { name: "Create" }).click();
    await expect(embedHost.frame.getByLabel("Notification")).toContainText(
      "Alert created",
    );

    await embedHost.navigate("alerts");
    await expect(
      embedHost.frame.getByRole("link", { name: ALICE_ALERT }),
    ).toBeVisible();
  });

  test("Should only show the alert to its creator and admins", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "bob" });
    await embedHost.navigate("alerts");
    await expect(
      embedHost.frame.getByText("You don't have any alerts yet"),
    ).toBeVisible();

    await embedHost.goto({ user: "admin" });
    await embedHost.navigate("alerts");
    await expect(
      embedHost.frame.getByRole("link", { name: ALICE_ALERT }),
    ).toBeVisible();
  });

  test("Should send emails when the alert triggers", async () => {
    test.setTimeout(60_000);
    const alice = getEmbedUser("alice");
    // The alert triggers when it's created, so look for emails from the start of the run.
    const since = new Date(RUN_ID);

    // Only the creator's email has an open link, and it points to the embed host's open page
    const aliceEmail = await getCompleteEmail(
      await waitForEmail(ALICE_ALERT, since, alice.email),
    );
    aliceOpenLink = extractOpenLink(aliceEmail) ?? "";
    expect(aliceOpenLink).toMatch(
      new RegExp(`^${getEmbedHostConfig().origin}/open\\.html\\?rill_open=`),
    );

    const extraEmail = await getCompleteEmail(
      await waitForEmail(ALICE_ALERT, since, getExtraRecipientEmail()),
    );
    expect(extractOpenLink(extraEmail)).toBeUndefined();

    // Only the extra recipient has an unsubscribe link, since the creator can't unsubscribe from their own alert
    expect(aliceEmail.Text).not.toContain("To unsubscribe from this alert");
    expect(extraEmail.Text).toContain("To unsubscribe from this alert");
  });

  test("Should open the alert as its creator", async ({ embedHost }) => {
    await embedHost.openLink(aliceOpenLink, "alice");

    // Lands on the alert's explore dashboard
    await expect(
      embedHost.frame.getByLabel("Open dimension details").first(),
    ).toBeVisible();
    await expect(
      embedHost.frame.getByRole("link", { name: "Home" }),
    ).toBeVisible();
  });

  test("Should not open the alert as another user", async ({ embedHost }) => {
    await embedHost.openLink(aliceOpenLink, "bob");

    await embedHost.expectError(
      "was not created by the provided external_user_id",
    );
    await expect(embedHost.page.locator("#rill-frame")).toBeHidden();
  });

  test("Should edit the alert", async ({ embedHost }) => {
    const alice = getEmbedUser("alice");

    await gotoAlert(embedHost, ALICE_ALERT);
    await embedHost.frame.getByRole("button", { name: "Edit alert" }).click();

    const alertForm = embedHost.frame.locator("form#edit-alert-form");
    await alertForm.getByRole("button", { name: "Next" }).click();
    await alertForm.getByRole("button", { name: "Next" }).click();
    // The saved creator email is kept
    await expect(alertForm.getByLabel("Your email")).toHaveValue(alice.email);

    await alertForm.getByTitle("Alert name").fill(ALICE_ALERT_EDITED);
    await alertForm.getByRole("button", { name: "Update" }).click();

    await expect(embedHost.frame.getByLabel("Notification")).toContainText(
      "Alert edited",
    );
    await expect(embedHost.frame.getByLabel("Alert name")).toHaveText(
      ALICE_ALERT_EDITED,
    );
  });

  test("Should delete the alert", async ({ embedHost }) => {
    await gotoAlert(embedHost, ALICE_ALERT_EDITED);
    await embedHost.frame.getByLabel("Alert context menu").click();
    await embedHost.frame
      .getByRole("menuitem", { name: "Delete Alert" })
      .click();

    await expect(
      embedHost.frame.getByText("You don't have any alerts yet"),
    ).toBeVisible();
  });
});

async function openAlertsMenu(embedHost: EmbedHostApp) {
  await embedHost.frame
    .getByRole("button", { name: "Alerts", exact: true })
    .click();
}

async function gotoAlert(embedHost: EmbedHostApp, title: string) {
  await embedHost.goto({ user: "alice" });
  await embedHost.navigate("alerts");
  await embedHost.frame.getByRole("link", { name: title }).click();
  await expect(embedHost.frame.getByLabel("Alert name")).toHaveText(title);
}
