import type { FrameLocator, Page } from "@playwright/test";

export type CanvasWidget = "KPI" | "Leaderboard" | "Text/Markdown";

/**
 * Adds a widget to the canvas open in the editor, through the "Add widget" button.
 * Takes a frame locator too, so it works for canvases in an embed's iframe.
 */
export async function addCanvasWidget(
  target: Page | FrameLocator,
  widget: CanvasWidget,
) {
  const addWidgetButton = target.getByRole("button", { name: "Add widget" });
  await addWidgetButton.waitFor({ state: "visible" });
  await addWidgetButton.click();
  await target.getByRole("menuitem", { name: widget, exact: true }).click();
}
