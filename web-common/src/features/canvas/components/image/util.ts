import { defaultImageAlignment } from "@rilldata/web-common/features/canvas/components/image";
import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";

// Return object-position CSS property for image
export function getImagePosition(alignment: ComponentAlignment | undefined) {
  if (!alignment) alignment = defaultImageAlignment;

  const verticalValue =
    alignment.vertical === "middle" ? "center" : alignment.vertical;

  const horizontalValue = alignment.horizontal;

  return `${horizontalValue} ${verticalValue}`;
}
