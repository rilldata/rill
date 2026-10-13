import { expect, test, type Page } from "@playwright/test";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { fakeHostPage, type FakeHostOptions } from "./fake-host";

const ORIGIN = "https://rill-widget.test";
const distPath = fileURLToPath(new URL("../dist/widget.html", import.meta.url));
const fixturesDir = fileURLToPath(
  new URL("../harness/fixtures/", import.meta.url),
);

interface Fixture {
  name: string;
  expect: "canvas" | "empty" | "error";
  arguments: Record<string, unknown>;
  result: { structuredContent?: Record<string, unknown> };
}

function loadFixture(name: string): Fixture {
  return JSON.parse(
    readFileSync(`${fixturesDir}${name}.json`, "utf8"),
  ) as Fixture;
}

function widgetHtml(): string {
  if (!existsSync(distPath)) {
    throw new Error(
      "dist/widget.html is missing; run `npm run build -w web-mcp-app` before the e2e tests",
    );
  }
  return readFileSync(distPath, "utf8");
}

async function mountWidget(page: Page, options: FakeHostOptions) {
  const errors: string[] = [];
  page.on("console", (message) => {
    if (message.type() === "error") errors.push(message.text());
  });
  page.on("pageerror", (error) => errors.push(error.message));
  const html = widgetHtml();
  await page.route(`${ORIGIN}/**`, async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (path === "/widget.html") {
      return route.fulfill({ contentType: "text/html", body: html });
    }
    if (path === "/") {
      return route.fulfill({
        contentType: "text/html",
        body: fakeHostPage(options),
      });
    }
    return route.fulfill({ status: 404, body: "not found" });
  });
  await page.goto(`${ORIGIN}/`);
  return { frame: page.frameLocator("#widget"), errors };
}

function sentMessages(page: Page, method: string) {
  return page.evaluate(
    (name) => window.__messages.filter((message) => message.method === name),
    method,
  );
}

const canonical = loadFixture("bar_chart");
const canonicalTitle = canonical.result.structuredContent!.title as string;
const openUrl = canonical.result.structuredContent!.open_url as string;

test("renders the canonical bar chart with title, size report and no console errors", async ({
  page,
}) => {
  const { frame, errors } = await mountWidget(page, { fixture: canonical });

  await expect(frame.getByText(canonicalTitle)).toBeVisible();
  const canvas = frame.locator("canvas").first();
  await expect(canvas).toBeVisible({ timeout: 15_000 });
  const box = await canvas.boundingBox();
  expect(box?.width ?? 0).toBeGreaterThan(0);
  expect(box?.height ?? 0).toBeGreaterThan(0);

  await expect
    .poll(() => sentMessages(page, "ui/notifications/size-changed"))
    .not.toHaveLength(0);
  const sizes = (await sentMessages(
    page,
    "ui/notifications/size-changed",
  )) as Array<{
    params: { height: number };
  }>;
  expect(sizes.at(-1)?.params.height ?? 0).toBeGreaterThan(100);

  await expect(
    frame.getByText("Sep 9 – Sep 15, 2023", { exact: false }),
  ).toBeVisible();
  expect(errors).toEqual([]);
});

test("has no horizontal overflow at 375px", async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 720 });
  const { frame, errors } = await mountWidget(page, {
    fixture: canonical,
    width: 351,
  });
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });
  const overflow = await frame
    .locator("html")
    .evaluate((root) => root.scrollWidth - root.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
  expect(errors).toEqual([]);
});

test("follows the host theme and style variables", async ({ page }) => {
  const { frame } = await mountWidget(page, {
    fixture: canonical,
    theme: "light",
    styles: {
      "--color-text-primary": "rgb(12, 34, 56)",
      "--font-sans": "Georgia, serif",
    },
  });
  const root = frame.locator("html");
  await expect(root).not.toHaveClass(/dark/);
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });
  // Tailwind text colors go through color-mix(), so compare the resolved tokens instead.
  const tokens = await root.evaluate((el) => {
    const style = getComputedStyle(el);
    return {
      host: style.getPropertyValue("--color-text-primary").trim(),
      rill: style.getPropertyValue("--fg-primary").trim(),
      font: style.getPropertyValue("--font-sans").trim(),
    };
  });
  expect(tokens.host).toBe("rgb(12, 34, 56)");
  expect(tokens.rill).toBe("rgb(12, 34, 56)");
  expect(tokens.font).toBe("Georgia, serif");
  const titleFont = await frame
    .getByText(canonicalTitle)
    .evaluate((el) => getComputedStyle(el).fontFamily);
  expect(titleFont).toBe("Georgia, serif");

  await page.evaluate(() => window.__host.setTheme("dark"));
  await expect(root).toHaveClass(/dark/);
  await expect(root).toHaveAttribute("data-theme", "dark");
  await expect(frame.locator("canvas").first()).toBeVisible();
});

test("requests fullscreen and shows the data table", async ({ page }) => {
  const { frame, errors } = await mountWidget(page, { fixture: canonical });
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });

  await frame.getByRole("button", { name: "Fullscreen" }).click();
  await expect
    .poll(() => sentMessages(page, "ui/request-display-mode"))
    .toEqual([expect.objectContaining({ params: { mode: "fullscreen" } })]);

  const dataTab = frame.getByRole("radio", { name: "Data" });
  await expect(dataTab).toBeVisible();
  await dataTab.click();
  await expect(
    frame.getByRole("columnheader", { name: "Advertiser Name" }),
  ).toBeVisible();
  await expect(frame.getByRole("cell", { name: "Acme Media" })).toBeVisible();

  // Interactive controls are offered because the fake host advertises serverTools.
  // The top-N control starts on the limit the result's spec carries (the backend applied 10).
  const topValues = frame.getByRole("radiogroup", { name: "Top values" });
  await expect(topValues).toBeVisible();
  await expect(
    topValues.getByRole("radio", { name: "10", exact: true }),
  ).toHaveAttribute("aria-checked", "true");
  await topValues.getByRole("radio", { name: "5", exact: true }).click();
  // Without a `project` in the tool input, none is forwarded (the runtime's own MCP server).
  await expect
    .poll(() => sentMessages(page, "tools/call"))
    .toEqual([
      expect.objectContaining({
        params: expect.objectContaining({
          name: "chart_data",
          arguments: { chart_type: "bar_chart", spec: expect.any(Object) },
        }),
      }),
    ]);
  // The model is told every control value, not only the ones the chart offers.
  await expect
    .poll(() => sentMessages(page, "ui/update-model-context"), {
      timeout: 5_000,
    })
    .toEqual([
      expect.objectContaining({
        params: expect.objectContaining({
          content: [
            {
              type: "text",
              text: expect.stringMatching(
                /^Chart now shows Total Bids per Advertiser Name: time grain n\/a, top 5 values, comparison off, Sep 9 – Sep 15, 2023\.$/,
              ),
            },
          ],
        }),
      }),
    ]);

  await frame.getByRole("button", { name: "Exit fullscreen" }).click();
  await expect(frame.getByRole("button", { name: "Fullscreen" })).toBeVisible();
  expect(errors).toEqual([]);
});

test("forwards the project of a unified MCP server call to chart_data", async ({
  page,
}) => {
  // Rill Cloud's unified server routes every tool call by a `project` argument.
  const { frame } = await mountWidget(page, {
    fixture: {
      ...canonical,
      arguments: { ...canonical.arguments, project: "demo/bids" },
    },
  });
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });
  await frame.getByRole("button", { name: "Fullscreen" }).click();
  await frame.getByRole("radio", { name: "5" }).click();
  await expect
    .poll(() => sentMessages(page, "tools/call"))
    .toEqual([
      expect.objectContaining({
        params: expect.objectContaining({
          name: "chart_data",
          arguments: {
            project: "demo/bids",
            chart_type: "bar_chart",
            spec: expect.any(Object),
          },
        }),
      }),
    ]);
});

test("offers the result's top-N limit when it is not a preset", async ({
  page,
}) => {
  // The funnel's stage channel carries the limit the backend applied (15), so the model is
  // told the real value and the control adds it to the presets.
  const { frame, errors } = await mountWidget(page, {
    fixture: loadFixture("funnel_chart"),
  });
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });
  await frame.getByRole("button", { name: "Fullscreen" }).click();
  const topValues = frame.getByRole("radiogroup", { name: "Top values" });
  await expect(topValues.getByRole("radio")).toHaveText([
    "5",
    "10",
    "15",
    "20",
  ]);
  await expect(
    topValues.getByRole("radio", { name: "15", exact: true }),
  ).toHaveAttribute("aria-checked", "true");

  await topValues.getByRole("radio", { name: "10", exact: true }).click();
  await expect
    .poll(() => sentMessages(page, "tools/call"))
    .toEqual([
      expect.objectContaining({
        params: expect.objectContaining({
          name: "chart_data",
          arguments: {
            chart_type: "funnel_chart",
            spec: expect.objectContaining({
              stage: expect.objectContaining({ limit: 10 }),
            }),
          },
        }),
      }),
    ]);
  await expect
    .poll(() => sentMessages(page, "ui/update-model-context"), {
      timeout: 5_000,
    })
    .toEqual([
      expect.objectContaining({
        params: expect.objectContaining({
          content: [
            {
              type: "text",
              text: "Chart now shows Total Bids funnel by Campaign Stage: time grain n/a, top 10 values, comparison off, Sep 9 – Sep 15, 2023.",
            },
          ],
        }),
      }),
    ]);
  expect(errors).toEqual([]);
});

test("opens the deep link through the host", async ({ page }) => {
  const { frame } = await mountWidget(page, { fixture: canonical });
  await frame.getByRole("button", { name: "Open in Rill" }).click();
  await expect
    .poll(() => sentMessages(page, "ui/open-link"))
    .toEqual([expect.objectContaining({ params: { url: openUrl } })]);
});

test("shows an error with Open in Rill when the result has no data", async ({
  page,
}) => {
  const { frame } = await mountWidget(page, {
    fixture: loadFixture("no_data"),
  });
  await expect(
    frame.getByText("did not include chart data", { exact: false }),
  ).toBeVisible();
  await expect(
    frame.getByRole("button", { name: "Open in Rill" }),
  ).toBeVisible();
  await expect(frame.locator("canvas")).toHaveCount(0);
});

test("falls back to an error for an unsupported widget mode", async ({
  page,
}) => {
  const { frame } = await mountWidget(page, {
    fixture: loadFixture("unsupported_widget"),
  });
  await expect(
    frame.getByText("needs a newer version", { exact: false }),
  ).toBeVisible();
  await expect(
    frame.getByRole("button", { name: "Open in Rill" }),
  ).toBeVisible();
});

test("shows the tool error text", async ({ page }) => {
  const { frame } = await mountWidget(page, {
    fixture: loadFixture("tool_error"),
  });
  await expect(frame.getByText("not found", { exact: false })).toBeVisible();
});

for (const [name, note] of [
  ["truncated", "Showing the top 10 values."],
  ["truncated_rows", "Showing the most recent periods."],
] as const) {
  test(`shows the truncation note for the ${name} fixture`, async ({
    page,
  }) => {
    const { frame } = await mountWidget(page, { fixture: loadFixture(name) });
    const banner = frame.getByRole("note");
    await expect(banner).toHaveText(note);
    // The server's own explanation stays available as the tooltip.
    await expect(banner).toHaveAttribute(
      "title",
      /incomplete|Lower the x limit/,
    );
  });
}

test("renders the comparison caption", async ({ page }) => {
  const { frame } = await mountWidget(page, {
    fixture: loadFixture("comparison"),
  });
  await expect(frame.locator("canvas").first()).toBeVisible({
    timeout: 15_000,
  });
  await expect(
    frame.getByText("vs. previous period", { exact: false }),
  ).toBeVisible();
});

test("localizes the actions from the host locale", async ({ page }) => {
  const { frame } = await mountWidget(page, {
    fixture: canonical,
    locale: "es-ES",
  });
  await expect(
    frame.getByRole("button", { name: "Abrir en Rill" }),
  ).toBeVisible();
});

const renderable = readdirSync(fixturesDir)
  .filter((file) => file.endsWith(".json"))
  .map((file) => file.replace(".json", ""))
  .filter((name) => loadFixture(name).expect === "canvas");

for (const name of renderable) {
  test(`renders a canvas for the ${name} fixture`, async ({ page }) => {
    const { frame, errors } = await mountWidget(page, {
      fixture: loadFixture(name),
    });
    const canvas = frame.locator("canvas").first();
    await expect(canvas).toBeVisible({ timeout: 15_000 });
    const box = await canvas.boundingBox();
    expect(box?.width ?? 0).toBeGreaterThan(0);
    expect(errors).toEqual([]);
  });
}

const empty = readdirSync(fixturesDir)
  .filter((file) => file.endsWith(".json"))
  .map((file) => file.replace(".json", ""))
  .filter((name) => loadFixture(name).expect === "empty");

for (const name of empty) {
  test(`shows the empty state for the ${name} fixture`, async ({ page }) => {
    const { frame, errors } = await mountWidget(page, {
      fixture: loadFixture(name),
    });
    await expect(frame.getByText("No data to display")).toBeVisible();
    await expect(frame.locator("canvas")).toHaveCount(0);
    expect(errors).toEqual([]);
  });
}
