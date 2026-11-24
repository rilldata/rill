import { expect, type Locator, type Page } from "@playwright/test";

// Helpers for canvas blocks (tab groups and containers) in the visual editor, shared by the
// tab group and container specs.

const INSTANCE_ID = "default";

/**
 * Writes a project file through the runtime that serves the page (Rill Developer serves the
 * UI and the API on one port), so a test can start from a known canvas.
 */
export async function putProjectFile(page: Page, path: string, blob: string) {
  const origin = new URL(page.url()).origin;
  const response = await page.request.post(
    `${origin}/v1/instances/${INSTANCE_ID}/files/entry`,
    { data: { path, blob, create: true } },
  );
  expect(response.ok()).toBe(true);
}

/** Reads a project file through the runtime that serves the page. */
export async function readProjectFile(
  page: Page,
  path: string,
): Promise<string> {
  const origin = new URL(page.url()).origin;
  const response = await page.request.get(
    `${origin}/v1/instances/${INSTANCE_ID}/files/entry`,
    { params: { path } },
  );
  if (!response.ok()) return "";
  return ((await response.json()) as { blob?: string }).blob ?? "";
}

/**
 * Waits until the project file matches: the editor saves its edits asynchronously.
 * `expected` lists strings the file must contain, in order.
 */
export async function expectProjectFile(
  page: Page,
  path: string,
  expected: string[],
  notExpected: string[] = [],
) {
  await expect
    .poll(async () => {
      const yaml = await readProjectFile(page, path);
      let from = 0;
      for (const snippet of expected) {
        const at = yaml.indexOf(snippet, from);
        if (at === -1) return `missing ${JSON.stringify(snippet)} in:\n${yaml}`;
        from = at + snippet.length;
      }
      for (const snippet of notExpected) {
        if (yaml.includes(snippet))
          return `unexpected ${JSON.stringify(snippet)} in:\n${yaml}`;
      }
      return "ok";
    })
    .toBe("ok");
}

/** The editor region of the block at a top-level index. */
export function blockRegion(page: Page, index: number): Locator {
  return page.locator(`#block-region-${index}`);
}

/**
 * Drags with the mouse from the center of `source` to `target` (its top edge by default,
 * where the drop zone between two top-level blocks is). Drags start after a short press,
 * like a user's.
 */
export async function dragTo(
  page: Page,
  source: Locator,
  target: Locator,
  at: "top" | "center" = "top",
) {
  const from = await source.boundingBox();
  if (!from) throw new Error("drag source is not visible");
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2);
  await page.mouse.down();
  await page.waitForTimeout(250);
  await page.mouse.move(
    from.x + from.width / 2 + 15,
    from.y + from.height / 2 + 15,
    {
      steps: 5,
    },
  );
  const to = await target.boundingBox();
  if (!to) throw new Error("drag target is not visible");
  await page.mouse.move(
    to.x + to.width / 2,
    at === "top" ? to.y + 2 : to.y + to.height / 2,
    { steps: 25 },
  );
  await page.waitForTimeout(300);
  await page.mouse.up();
}

/** Opens the "Add widget" menu of the floating button and picks an item. */
export async function addFromFloatingMenu(page: Page, item: string) {
  await page.getByRole("button", { name: "Add widget" }).last().click();
  await page.getByRole("menuitem", { name: item, exact: true }).click();
}
