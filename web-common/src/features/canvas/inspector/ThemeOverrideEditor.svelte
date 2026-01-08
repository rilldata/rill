<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { getParsedDocument } from "./selectors";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { get } from "svelte/store";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    COMPONENT_PATH_ROW_INDEX,
    COMPONENT_PATH_COLUMN_INDEX,
  } from "../stores/canvas-entity";
  import {
    getComponentThemeOverrides,
    mergeComponentTheme,
  } from "../utils/component-colors";
  import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
  import { isChartComponentType } from "../components/util";

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

  // Get parsed document (needed for writing updates)
  $: parsedDocumentStore = getParsedDocument(fileArtifact);
  $: parsedDocument = get(parsedDocumentStore);

  // Get canvas spec to access item properties
  $: specStore = component.parent?.specStore;
  $: canvasData = specStore ? $specStore?.data : undefined;
  $: canvasRows = canvasData?.canvas?.rows ?? [];

  // Get the item for this component with bounds checking
  $: item = (() => {
    if (!isValidPosition || !canvasRows || rowIndex >= canvasRows.length) {
      return undefined;
    }
    const row = canvasRows[rowIndex];
    if (!row?.items || columnIndex >= row.items.length) {
      return undefined;
    }
    return row.items[columnIndex];
  })();

  // Get component theme overrides
  $: themeOverrides = getComponentThemeOverrides(item);

  // Get global theme for reference colors
  $: globalTheme = component.parent?.theme;
  $: isDarkMode =
    typeof window !== "undefined" ? $themeControl === "dark" : false;
  $: globalThemeObject = themeManager.resolveThemeObject(
    $globalTheme?.spec,
    isDarkMode,
  );

  // Check if this is a chart component
  $: isChart = isChartComponentType(component.type);

  // Common theme variables to show in UI
  const commonThemeVars = [
    "primary",
    "secondary",
    "card",
    "foreground",
    "border",
  ] as const;

  type ThemeVar = (typeof commonThemeVars)[number];

  // Get display value for a theme variable (override or global theme or fallback)
  function getDisplayValue(varName: ThemeVar, mode: "light" | "dark"): string {
    const override =
      mode === "light" ? themeOverrides.light : themeOverrides.dark;

    // Check if override exists
    if (override) {
      if (varName === "primary" && override.primary) {
        return override.primary;
      }
      if (varName === "secondary" && override.secondary) {
        return override.secondary;
      }
      if (override.variables?.[varName]) {
        return override.variables[varName];
      }
    }

    // Fallback to global theme
    if (globalThemeObject?.[varName]) {
      return globalThemeObject[varName];
    }

    // Ultimate fallback
    const fallbacks: Record<ThemeVar, string> = {
      primary: "#4F46E5",
      secondary: "#8B5CF6",
      card: "#ffffff",
      foreground: "#000000",
      border: "#e5e7eb",
    };
    return fallbacks[varName];
  }

  // Check if a variable has an override
  function hasOverride(varName: ThemeVar, mode: "light" | "dark"): boolean {
    const override =
      mode === "light" ? themeOverrides.light : themeOverrides.dark;
    if (!override) return false;

    if (varName === "primary") return !!override.primary;
    if (varName === "secondary") return !!override.secondary;
    return !!override.variables?.[varName];
  }

  // Debounced version of updateThemeVar to prevent spam
  const debouncedUpdateThemeVar = debounce(
    async (
      varName: ThemeVar,
      mode: "light" | "dark",
      color: string | undefined,
    ) => {
      await updateThemeVarInternal(varName, mode, color);
    },
    300,
  );

  async function updateThemeVarInternal(
    varName: ThemeVar,
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

      // Get current theme_override structure (if exists)
      const currentThemeOverride = item.get("theme_override");
      const currentModeNode = currentThemeOverride?.get(mode);

      // Build new mode node - variables should be at the same level as primary/secondary
      let newModeNode: Record<string, unknown> = {};

      if (currentModeNode) {
        // Preserve existing values by converting the entire node to JSON
        const currentModeData = currentModeNode.toJSON() as Record<
          string,
          unknown
        >;

        // Copy all existing values - variables are at the same level as primary/secondary
        for (const [key, value] of Object.entries(currentModeData)) {
          if (key !== "variables") {
            newModeNode[key] = value;
          }
        }
      }

      if (!color || color.trim() === "") {
        // Remove the override - variables are at the same level as primary/secondary
        delete newModeNode[varName];

        // If mode node is empty, remove it
        if (Object.keys(newModeNode).length === 0) {
          // Remove the mode from theme_override
          const newThemeOverride: Record<string, unknown> = {};
          if (currentThemeOverride) {
            const otherMode = mode === "light" ? "dark" : "light";
            const otherModeNode = currentThemeOverride.get(otherMode);
            if (otherModeNode) {
              newThemeOverride[otherMode] = otherModeNode.toJS();
            }
          }

          // If theme_override is empty, remove it
          if (Object.keys(newThemeOverride).length === 0) {
            doc.deleteIn([
              "rows",
              rowIndex,
              "items",
              columnIndex,
              "theme_override",
            ]);
          } else {
            doc.setIn(
              ["rows", rowIndex, "items", columnIndex, "theme_override"],
              newThemeOverride,
            );
          }
        } else {
          // Update theme_override with new mode node
          const newThemeOverride: Record<string, unknown> = {
            [mode]: newModeNode,
          };
          if (currentThemeOverride) {
            const otherMode = mode === "light" ? "dark" : "light";
            const otherModeNode = currentThemeOverride.get(otherMode);
            if (otherModeNode) {
              newThemeOverride[otherMode] = otherModeNode.toJS();
            }
          }
          doc.setIn(
            ["rows", rowIndex, "items", columnIndex, "theme_override"],
            newThemeOverride,
          );
        }
      } else {
        // Set the override (trim whitespace) - variables are at the same level as primary/secondary
        const trimmedColor = color.trim();
        newModeNode[varName] = trimmedColor;

        // Update theme_override with new mode node
        const newThemeOverride: Record<string, unknown> = {
          [mode]: newModeNode,
        };
        if (currentThemeOverride) {
          const otherMode = mode === "light" ? "dark" : "light";
          const otherModeNode = currentThemeOverride.get(otherMode);
          if (otherModeNode) {
            newThemeOverride[otherMode] = otherModeNode.toJSON();
          }
        }
        doc.setIn(
          ["rows", rowIndex, "items", columnIndex, "theme_override"],
          newThemeOverride,
        );
      }

      updateEditorContent(doc.toString(), false, true);
      await saveLocalContent();
      // Clear error on success
      errorMessage = null;
    } catch (e) {
      errorMessage = `Failed to update ${varName} color: ${e}`;
      console.error(`${varName} color update error:`, e);
    }
  }

  function handleColorChange(
    varName: ThemeVar,
    mode: "light" | "dark",
    color: string,
  ) {
    debouncedUpdateThemeVar(varName, mode, color);
  }

  function resetToTheme(varName: ThemeVar, mode: "light" | "dark") {
    // Reset should be immediate, not debounced
    updateThemeVarInternal(varName, mode, undefined);
  }
</script>

{#if !isValidPosition}
  <div class="component-param">
    <div class="error-message">
      <p class="text-sm text-red-600">
        Cannot edit theme overrides: Invalid component position
      </p>
    </div>
  </div>
{:else if errorMessage}
  <div class="component-param">
    <div class="error-message">
      <p class="text-sm text-red-600">{errorMessage}</p>
    </div>
  </div>
{:else}
  <div class="component-param">
    <InputLabel
      small
      label="Theme Overrides"
      id="component-theme-overrides"
      hint="Override theme variables for this component. Changes apply to both light and dark modes separately."
    />

    <div class="flex flex-col gap-y-4 mt-4">
      {#each commonThemeVars as varName}
        <div class="flex flex-col gap-y-2">
          <InputLabel
            small
            label={varName.charAt(0).toUpperCase() + varName.slice(1)}
            id="component-theme-{varName}"
            hint={varName === "foreground" && isChart
              ? "Sets the color for chart axis labels and titles"
              : undefined}
          />
          <div class="flex flex-col gap-y-2">
            <!-- Light mode -->
            <div class="flex items-center justify-between">
              <ColorInput
                small
                stringColor={getDisplayValue(varName, "light")}
                label="Light mode"
                labelFirst
                allowLightnessControl
                onChange={(color) => handleColorChange(varName, "light", color)}
              />
              {#if hasOverride(varName, "light")}
                <Button
                  type="text"
                  size="sm"
                  onClick={() => resetToTheme(varName, "light")}
                  title="Reset to theme default"
                >
                  Reset
                </Button>
              {/if}
            </div>
            <!-- Dark mode -->
            <div class="flex items-center justify-between">
              <ColorInput
                small
                stringColor={getDisplayValue(varName, "dark")}
                label="Dark mode"
                labelFirst
                allowLightnessControl
                onChange={(color) => handleColorChange(varName, "dark", color)}
              />
              {#if hasOverride(varName, "dark")}
                <Button
                  type="text"
                  size="sm"
                  onClick={() => resetToTheme(varName, "dark")}
                  title="Reset to theme default"
                >
                  Reset
                </Button>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

<style lang="postcss">
  .component-param {
    @apply py-3 px-5 border-t;
  }

  .error-message {
    @apply mt-2;
  }
</style>

