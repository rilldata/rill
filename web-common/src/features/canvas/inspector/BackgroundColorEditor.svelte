<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { getParsedDocument } from "./selectors";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { get } from "svelte/store";
  import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import {
    COMPONENT_PATH_ROW_INDEX,
    COMPONENT_PATH_COLUMN_INDEX,
  } from "../stores/canvas-entity";

  export let component: BaseCanvasComponent;
  export let fileArtifact: FileArtifact;

  let errorMessage: string | null = null;

  // Get component position from pathInYAML with bounds checking
  $: rowIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_ROW_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  $: columnIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_COLUMN_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  // Validate indices before proceeding
  $: isValidPosition = rowIndex >= 0 && columnIndex >= 0;

  // Get parsed document
  $: parsedDocumentStore = getParsedDocument(fileArtifact);
  $: parsedDocument = get(parsedDocumentStore);

  // Get current background colors from YAML with proper error handling
  $: currentLightColor = (() => {
    if (!isValidPosition) return undefined;
    try {
      const value = parsedDocument.getIn([
        "rows",
        rowIndex,
        "items",
        columnIndex,
        "background_color_light",
      ]) as string | undefined;
      return typeof value === "string" && value.trim() !== ""
        ? value.trim()
        : undefined;
    } catch (e) {
      errorMessage = `Failed to read light mode color: ${e}`;
      return undefined;
    }
  })();

  $: currentDarkColor = (() => {
    if (!isValidPosition) return undefined;
    try {
      const value = parsedDocument.getIn([
        "rows",
        rowIndex,
        "items",
        columnIndex,
        "background_color_dark",
      ]) as string | undefined;
      return typeof value === "string" && value.trim() !== ""
        ? value.trim()
        : undefined;
    } catch (e) {
      errorMessage = `Failed to read dark mode color: ${e}`;
      return undefined;
    }
  })();

  // Get theme's card color as reference (SSR-safe)
  $: isDarkMode =
    typeof window !== "undefined" ? $themeControl === "dark" : false;
  $: themeCardColor = (() => {
    if (typeof window === "undefined") return undefined;
    const color = themeManager.resolveCSSVariable("var(--card)", isDarkMode);
    // If resolved color is still a CSS variable, return undefined
    return color && !color.startsWith("var(") ? color : undefined;
  })();

  // Display colors (use current override or theme color)
  $: displayLightColor = currentLightColor || themeCardColor || "#ffffff";
  $: displayDarkColor = currentDarkColor || themeCardColor || "#1a1a1a";

  async function updateBackgroundColor(
    mode: "light" | "dark",
    color: string | undefined,
  ) {
    if (!isValidPosition) {
      errorMessage = "Invalid component position";
      return;
    }

    errorMessage = null;
    const { updateEditorContent, saveLocalContent } = fileArtifact;

    try {
      const doc = get(parsedDocumentStore);
      const path = [
        "rows",
        rowIndex,
        "items",
        columnIndex,
        `background_color_${mode}`,
      ];

      // Validate path exists
      const rows = doc.get("rows");
      if (!rows) {
        throw new Error("Canvas has no rows");
      }

      const row = rows.get(rowIndex);
      if (!row) {
        throw new Error(`Row ${rowIndex} does not exist`);
      }

      const items = row.get("items");
      if (!items) {
        throw new Error(`Row ${rowIndex} has no items`);
      }

      const item = items.get(columnIndex);
      if (!item) {
        throw new Error(
          `Item ${columnIndex} does not exist in row ${rowIndex}`,
        );
      }

      if (!color || color.trim() === "") {
        // Remove the override if color is empty
        doc.deleteIn(path);
      } else {
        // Set the override (trim whitespace)
        const trimmedColor = color.trim();
        doc.setIn(path, trimmedColor);
      }

      updateEditorContent(doc.toString(), false, true);
      await saveLocalContent();
    } catch (e) {
      errorMessage = `Failed to update background color: ${e}`;
      console.error("Background color update error:", e);
    }
  }

  function handleLightColorChange(color: string) {
    updateBackgroundColor("light", color);
  }

  function handleDarkColorChange(color: string) {
    updateBackgroundColor("dark", color);
  }

  function resetToTheme(mode: "light" | "dark") {
    updateBackgroundColor(mode, undefined);
  }
</script>

<div class="component-param">
  <InputLabel
    small
    label="Background color"
    id="component-background-color"
    hint="Override theme's card color for this component"
  />
  {#if !isValidPosition}
    <div class="error-message">
      <p class="text-sm text-red-600">
        Cannot edit background color: Invalid component position
      </p>
    </div>
  {:else if errorMessage}
    <div class="error-message">
      <p class="text-sm text-red-600">{errorMessage}</p>
    </div>
  {:else}
    <div class="flex flex-col gap-y-2 mt-2">
      <div class="flex flex-col gap-y-1">
        <div class="flex items-center justify-between">
          <ColorInput
            small
            stringColor={displayLightColor}
            label="Light mode"
            labelFirst
            allowLightnessControl
            onChange={handleLightColorChange}
          />
          {#if currentLightColor}
            <Button
              type="text"
              size="sm"
              onClick={() => resetToTheme("light")}
              title="Reset to theme default"
            >
              Reset
            </Button>
          {/if}
        </div>
      </div>
      <div class="flex flex-col gap-y-1">
        <div class="flex items-center justify-between">
          <ColorInput
            small
            stringColor={displayDarkColor}
            label="Dark mode"
            labelFirst
            allowLightnessControl
            onChange={handleDarkColorChange}
          />
          {#if currentDarkColor}
            <Button
              type="text"
              size="sm"
              onClick={() => resetToTheme("dark")}
              title="Reset to theme default"
            >
              Reset
            </Button>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5 border-t;
  }

  .error-message {
    @apply mt-2;
  }
</style>
