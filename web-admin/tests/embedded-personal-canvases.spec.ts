import { expect } from "@playwright/test";
import { addCanvasWidget } from "@rilldata/web-common/tests/utils/canvas-helpers";
import type { EmbedHostApp } from "./embed-host/embed-host-app.ts";
import { test } from "./setup/base";

// Personal canvases are enabled for embed users through the `personal_canvases` feature flag in the AdBids project's `rill.yaml`.
test.describe("Embedded personal canvases", () => {
  test("Should be hidden for embeds without an external user ID", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "none" });
    await expectEmbedLoaded(embedHost);

    await expect(myDashboardsButton(embedHost)).toHaveCount(0);
  });

  test("Should be hidden for embeds without navigation", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "alice", navigation: false });
    await expectEmbedLoaded(embedHost);
    await expect(myDashboardsButton(embedHost)).toHaveCount(0);

    await embedHost.setNavigation(true);
    await expect(myDashboardsButton(embedHost)).toBeVisible();
  });

  test("Should only be accessible to the embed user who created them", async ({
    embedHost,
  }) => {
    const { frame } = embedHost;
    const displayName = `Personal canvas ${Date.now()}`;

    // Alice creates a personal canvas with a KPI and a leaderboard.
    await embedHost.goto({ user: "alice" });
    await myDashboardsButton(embedHost).click();
    await frame.getByRole("menuitem", { name: "New dashboard" }).click();
    await frame.getByLabel("Display name").fill(displayName);
    await frame.getByRole("button", { name: "Create", exact: true }).click();

    await addCanvasWidget(frame, "KPI");
    await expect(
      frame.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();
    await addCanvasWidget(frame, "Leaderboard");
    await expect(frame.getByLabel(/ leaderboard$/).first()).toBeVisible();

    // The previewed canvas shows the saved components.
    await frame.getByRole("button", { name: "Preview" }).click();
    await expect(
      frame.getByRole("heading", { name: displayName }),
    ).toBeVisible();
    await expect(
      frame.getByRole("heading", { name: "Sum of Bid Price" }),
    ).toBeVisible();
    await expect(frame.getByLabel(/ leaderboard$/).first()).toBeVisible();

    // Alice sees the canvas on the home page and in the menu.
    await embedHost.navigate("home");
    await expectHomeLoaded(embedHost);
    await expect(
      frame.getByRole("link").filter({ hasText: displayName }),
    ).toBeVisible();
    await myDashboardsButton(embedHost).click();
    await expect(
      frame.getByRole("menuitem").filter({ hasText: displayName }),
    ).toBeVisible();
    await embedHost.page.keyboard.press("Escape");

    // Bob doesn't see the canvas.
    await embedHost.goto({ user: "bob" });
    await embedHost.navigate("home");
    await expectHomeLoaded(embedHost);
    await expect(frame.getByText(displayName)).toHaveCount(0);
    await myDashboardsButton(embedHost).click();
    await expect(frame.getByText("No personal dashboards yet")).toBeVisible();
    await expect(
      frame.getByRole("menuitem").filter({ hasText: displayName }),
    ).toHaveCount(0);
    await embedHost.page.keyboard.press("Escape");

    // Alice deletes the canvas.
    await embedHost.goto({ user: "alice" });
    await myDashboardsButton(embedHost).click();
    await frame
      .getByRole("button", { name: `Delete ${displayName}`, exact: true })
      .click();
    const dialog = frame.getByRole("alertdialog");
    await dialog.getByLabel("confirmation").fill(`delete ${displayName}`);
    await dialog.getByRole("button", { name: "Delete", exact: true }).click();
    await expect(dialog).toBeHidden();

    await myDashboardsButton(embedHost).click();
    await expect(
      frame.getByRole("menuitem").filter({ hasText: displayName }),
    ).toHaveCount(0);
  });
});

function myDashboardsButton(embedHost: EmbedHostApp) {
  return embedHost.frame.getByRole("button", { name: "My dashboards" });
}

/** Waits for the default explore dashboard to render. */
async function expectEmbedLoaded(embedHost: EmbedHostApp) {
  await embedHost.recorder.waitForReady();
  await expect(
    embedHost.frame.getByLabel(/ leaderboard$/).first(),
  ).toBeVisible();
}

/** Waits for the project's dashboards to be listed on the embed's home page. */
async function expectHomeLoaded(embedHost: EmbedHostApp) {
  await expect(
    embedHost.frame.getByText("Adbids Canvas Dashboard"),
  ).toBeVisible();
}
