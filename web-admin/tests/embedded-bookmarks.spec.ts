import { expect } from "@playwright/test";
import type { EmbedHostApp } from "./embed-host/embed-host-app.ts";
import { test } from "./setup/base";

// Bookmarks are enabled for embed users through the `embedded_bookmarks` feature flag in the AdBids project's `rill.yaml`.
test.describe("Embedded bookmarks", () => {
  test("Should be hidden for embeds without an external user ID", async ({
    embedHost,
  }) => {
    await embedHost.goto({ user: "none" });
    await expectEmbedLoaded(embedHost);

    await expect(bookmarksButton(embedHost)).toHaveCount(0);
  });

  test("Should only be accessible to the embed user who created them", async ({
    embedHost,
  }) => {
    const { frame } = embedHost;
    const displayName = `Explore bookmark ${Date.now()}`;
    const renamed = `${displayName} renamed`;

    // Alice filters the explore dashboard and bookmarks it.
    await embedHost.goto({ user: "alice" });
    await expectEmbedLoaded(embedHost);
    // Embeds only have personal bookmarks, so there is no home bookmark.
    await expect(
      frame.getByRole("button", { name: "Home bookmark dropdown" }),
    ).toHaveCount(0);
    await frame.getByRole("row", { name: /^Facebook/ }).click();
    await expect(publisherFilter(embedHost)).toBeVisible();

    await bookmarksButton(embedHost).click();
    // Embeds only have personal bookmarks, so there is no managed bookmarks section.
    await expect(frame.getByText("Managed bookmarks")).toHaveCount(0);
    await frame
      .getByRole("menuitem", { name: "Bookmark current view", exact: true })
      .click();
    // Embed users can't create managed bookmarks, so there is no category to pick.
    await expect(frame.getByLabel("Category")).toHaveCount(0);
    await frame.getByLabel("Label").fill(displayName);
    await frame.getByRole("button", { name: "Save" }).click();
    await expect(frame.getByText("Bookmark created")).toBeVisible();

    // Reloading the embed resets the dashboard; opening the bookmark applies the filter again.
    await embedHost.goto({ user: "alice" });
    await expectEmbedLoaded(embedHost);
    await expect(publisherFilter(embedHost)).toHaveCount(0);
    await bookmarksButton(embedHost).click();
    await bookmarkEntry(embedHost, displayName).getByRole("link").click();
    await expect(publisherFilter(embedHost)).toBeVisible();

    // Alice renames the bookmark.
    await bookmarksButton(embedHost).click();
    await bookmarkEntry(embedHost, displayName).hover();
    await bookmarkEntry(embedHost, displayName)
      .getByRole("button")
      .first()
      .click();
    await frame.getByLabel("Label").fill(renamed);
    await frame.getByRole("button", { name: "Save" }).click();
    await expect(frame.getByText("Bookmark updated")).toBeVisible();

    // Bob doesn't see the bookmark.
    await embedHost.goto({ user: "bob" });
    await expectEmbedLoaded(embedHost);
    await bookmarksButton(embedHost).click();
    await expect(
      frame.getByText("You have no bookmarks for this dashboard."),
    ).toBeVisible();
    await expect(bookmarkEntry(embedHost, renamed)).toHaveCount(0);
    await embedHost.page.keyboard.press("Escape");

    // Alice deletes the bookmark.
    await embedHost.goto({ user: "alice" });
    await expectEmbedLoaded(embedHost);
    await bookmarksButton(embedHost).click();
    await bookmarkEntry(embedHost, renamed).hover();
    await bookmarkEntry(embedHost, renamed)
      .getByRole("button", { name: "Delete bookmark" })
      .click();
    await expect(frame.getByText(`Bookmark ${renamed} deleted`)).toBeVisible();
    await bookmarksButton(embedHost).click();
    await expect(bookmarkEntry(embedHost, renamed)).toHaveCount(0);
  });

  test("Should bookmark canvas dashboards", async ({ embedHost }) => {
    const { frame } = embedHost;
    const displayName = `Canvas bookmark ${Date.now()}`;

    await embedHost.goto({ user: "alice" });
    await embedHost.navigate("home");
    await frame.getByRole("link", { name: /Adbids Canvas Dashboard/ }).click();
    await expect(bookmarksButton(embedHost)).toBeVisible();

    await bookmarksButton(embedHost).click();
    await frame
      .getByRole("menuitem", { name: "Bookmark current view", exact: true })
      .click();
    await frame.getByLabel("Label").fill(displayName);
    await frame.getByRole("button", { name: "Save" }).click();
    await expect(frame.getByText("Bookmark created")).toBeVisible();

    await bookmarksButton(embedHost).click();
    await bookmarkEntry(embedHost, displayName).hover();
    await bookmarkEntry(embedHost, displayName)
      .getByRole("button", { name: "Delete bookmark" })
      .click();
    await expect(
      frame.getByText(`Bookmark ${displayName} deleted`),
    ).toBeVisible();
  });
});

function bookmarksButton(embedHost: EmbedHostApp) {
  return embedHost.frame.getByRole("button", {
    name: "Other bookmark dropdown",
  });
}

function bookmarkEntry(embedHost: EmbedHostApp, displayName: string) {
  return embedHost.frame.getByLabel(`${displayName} Bookmark Entry`);
}

function publisherFilter(embedHost: EmbedHostApp) {
  return embedHost.frame.getByLabel(/^publisher filter$/i);
}

/** Waits for the default explore dashboard to render. */
async function expectEmbedLoaded(embedHost: EmbedHostApp) {
  await embedHost.recorder.waitForReady();
  await expect(
    embedHost.frame.getByLabel(/ leaderboard$/).first(),
  ).toBeVisible();
}
