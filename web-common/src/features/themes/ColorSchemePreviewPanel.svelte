<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { Theme } from "./theme";
  import PreviewComponents from "./PreviewComponents.svelte";

  export let filePath: string;

  const INITIAL_HEIGHT = 600;
  const MIN_HEIGHT = 300;
  const MAX_HEIGHT = 600;

  let isOpen = false;
  let height = INITIAL_HEIGHT;
  let previewMode: "light" | "dark" = "light";

  // Get file content and parse YAML to create Theme object
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ editorContent } = fileArtifact);

  $: parsedDocument = $editorContent ? parseDocument($editorContent) : null;
  $: themeData = (parsedDocument?.toJSON() || {}) as V1ThemeSpec;

  // Create Theme instance - this processes the YAML and generates CSS
  $: theme = themeData ? new Theme(themeData) : new Theme(undefined);

  // Extract colors directly from the parsed YAML (not from Theme class)
  // The YAML has colors at the root level of light/dark, not under variables
  $: lightModeYaml = (themeData?.light || {}) as Record<string, string>;
  $: darkModeYaml = (themeData?.dark || {}) as Record<string, string>;
  $: currentModeYaml = previewMode === "light" ? lightModeYaml : darkModeYaml;

  // Extract palette colors into arrays for easy access
  $: sequentialColors = Array.from({ length: 9 }, (_, i) => {
    const key = `color-sequential-${i + 1}`;
    return currentModeYaml[key] || `var(--color-sequential-${i + 1})`;
  });

  $: divergingColors = Array.from({ length: 11 }, (_, i) => {
    const key = `color-diverging-${i + 1}`;
    return currentModeYaml[key] || `var(--color-diverging-${i + 1})`;
  });

  $: qualitativeColors = Array.from({ length: 24 }, (_, i) => {
    const key = `color-qualitative-${i + 1}`;
    return currentModeYaml[key] || `var(--color-qualitative-${i + 1})`;
  });

  // Extract theme colors
  $: primaryColor = currentModeYaml["primary"] || "var(--color-theme-500)";
  $: backgroundColor = currentModeYaml["background"] || (previewMode === "light" ? "#f9fafb" : "#111827");
  $: cardColor = currentModeYaml["card"] || (previewMode === "light" ? "#ffffff" : "#374151");

  async function toggle() {
    isOpen = !isOpen;
  }

  function handleModeChange(_: number, value: string) {
    previewMode = value.toLowerCase() as "light" | "dark";
  }
</script>

<div
  class="relative w-full flex-none overflow-hidden flex flex-col bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700"
>
  <Resizer
    disabled={!isOpen}
    dimension={height}
    min={MIN_HEIGHT}
    max={MAX_HEIGHT}
    basis={INITIAL_HEIGHT}
    onUpdate={(dimension) => (height = dimension)}
    direction="NS"
  />
  <div class="bar">
    <button
      aria-label="Toggle color scheme preview"
      class="text-xs text-gray-800 dark:text-gray-200 rounded-sm hover:bg-gray-100 dark:hover:bg-gray-800 h-6 px-1.5 py-px flex items-center gap-1.5 transition-colors"
      on:click={toggle}
    >
      <span class="transition-transform" class:rotate-180={isOpen}>
        <CaretDownIcon size="14px" />
      </span>
      <span class="font-semibold">Color Scheme Preview</span>
    </button>
    {#if isOpen}
      <div class="ml-auto pt-1">
        <FieldSwitcher
          small
          fields={["Light", "Dark"]}
          selected={previewMode === "light" ? 0 : 1}
          onClick={handleModeChange}
        />
      </div>
    {/if}
  </div>

  {#if isOpen}
    <div
      class="overflow-y-auto px-4 py-4"
      class:dark={previewMode === "dark"}
      style="height: {height - 28}px; background-color: {backgroundColor};"
    >
      <!-- Wrap in ThemeProvider to apply the theme CSS -->
      <ThemeProvider {theme}>
        <PreviewComponents
          {sequentialColors}
          {qualitativeColors}
          {divergingColors}
          {primaryColor}
          {cardColor}
        />
      </ThemeProvider>
    </div>
  {/if}
</div>

<style lang="postcss">
  .bar {
    @apply flex items-center px-3 h-7 w-full bg-gray-50 border-b border-gray-200;
  }

  :global(.dark) .bar {
    @apply bg-gray-950 border-gray-700;
  }
</style>
