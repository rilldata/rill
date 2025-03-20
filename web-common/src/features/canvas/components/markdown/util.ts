import { defaultMarkdownAlignment } from "@rilldata/web-common/features/canvas/components/markdown";
import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";

export function getPositionClasses(alignment: ComponentAlignment | undefined) {
  if (!alignment) alignment = defaultMarkdownAlignment;
  let classString = "";

  switch (alignment.horizontal) {
    case "left":
      classString = "items-start";
      break;
    case "center":
      classString = "items-center";
      break;
    case "right":
      classString = "items-end";
  }

  switch (alignment.vertical) {
    case "top":
      classString += " justify-start";
      break;
    case "middle":
      classString += " justify-center";
      break;
    case "bottom":
      classString += " justify-end";
  }

  return classString;
}
