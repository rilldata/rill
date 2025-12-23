const DEFAULT_VIEWPORT_MARGIN_PX = 16;
const DEFAULT_MENU_WIDTH_PX = 240;

export function determineDropdownAlign({
  triggerRect,
  menuWidth,
  viewportWidth,
  boundaryRight,
  margin = DEFAULT_VIEWPORT_MARGIN_PX,
}: {
  triggerRect: DOMRect | null;
  menuWidth?: number;
  viewportWidth: number;
  boundaryRight?: number | null;
  margin?: number;
}): "start" | "end" {
  if (!triggerRect) return "start";

  const width =
    typeof menuWidth === "number" ? menuWidth : DEFAULT_MENU_WIDTH_PX;
  const comparisonRight =
    typeof boundaryRight === "number"
      ? Math.min(boundaryRight, viewportWidth)
      : viewportWidth;
  const projectedRightEdge = triggerRect.left + width;
  const safeRightEdge = comparisonRight - margin;

  return projectedRightEdge > safeRightEdge ? "end" : "start";
}

export const DROPDOWN_DEFAULT_MARGIN = DEFAULT_VIEWPORT_MARGIN_PX;
