const DEFAULT_VIEWPORT_MARGIN_PX = 16;
const DEFAULT_MENU_WIDTH_PX = 240;

export function determineDropdownAlign({
  triggerRect,
  menuWidth,
  viewportWidth,
  margin = DEFAULT_VIEWPORT_MARGIN_PX,
}: {
  triggerRect: DOMRect | null;
  menuWidth?: number;
  viewportWidth: number;
  margin?: number;
}): "start" | "end" {
  if (!triggerRect) return "start";

  const width =
    typeof menuWidth === "number" ? menuWidth : DEFAULT_MENU_WIDTH_PX;
  const projectedRightEdge = triggerRect.left + width;
  const safeViewportRight = viewportWidth - margin;

  return projectedRightEdge > safeViewportRight ? "end" : "start";
}

export const DROPDOWN_DEFAULT_MARGIN = DEFAULT_VIEWPORT_MARGIN_PX;
