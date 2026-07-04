import { type Document, YAMLMap } from "yaml";
import { writable } from "svelte/store";

/**
 * Shared store for theme preview mode across editor and inspector
 */
export const themePreviewMode = writable<PreviewMode>("light");

export type PreviewMode = "light" | "dark";

/**
 * Updates a color value in a parsed YAML theme document
 * @param parsedDocument - The parsed YAML document
 * @param mode - The theme mode ("light" or "dark")
 * @param colorKey - The color key to update (e.g., "primary", "surface-background")
 * @param value - The new color value
 * @param removeComments - Whether to remove inline comments from the edited line
 * @returns The updated YAML string
 */
export function updateThemeColor(
  parsedDocument: Document,
  mode: PreviewMode,
  colorKey: string,
  value: string,
  removeComments: boolean = true,
): string {
  // Get or create the mode section (light/dark)
  let modeSection: YAMLMap = parsedDocument.get(mode, true) as YAMLMap;

  // If the mode section doesn't exist or isn't a YAMLMap, create it
  if (!(modeSection instanceof YAMLMap)) {
    const newMap = new YAMLMap();
    parsedDocument.set(mode, newMap);
    modeSection = newMap;
  }

  // Set the color value
  modeSection.set(colorKey, value);

  // Optionally remove inline comment from the edited line
  if (removeComments) {
    const valueNode = modeSection.get(colorKey, true);
    if (
      valueNode &&
      typeof valueNode === "object" &&
      valueNode !== null &&
      "comment" in valueNode
    ) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (valueNode as any).comment = undefined;
    }
  }

  return parsedDocument.toString();
}
