<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import PreviewComponents from "./PreviewComponents.svelte";
  import {
    parseThemeFromYaml,
    extractThemeColors,
    type PreviewMode,
  } from "./theme-preview-utils";

  export let filePath: string;

  const INITIAL_HEIGHT = 380;
  const MIN_HEIGHT = 180;
  const MAX_HEIGHT = 500;

  let isOpen = false;
  let height = INITIAL_HEIGHT;
  let previewMode: PreviewMode = "light";

  // Get file content and parse YAML to create Theme object
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ editorContent } = fileArtifact);

  // Parse theme and extract colors using shared utility
  $: ({ theme, themeData } = parseThemeFromYaml($editorContent));
  $: ({
    sequentialColors,
    divergingColors,
    qualitativeColors,
    primaryColor,
    backgroundColor,
    cardColor,
    fgPrimary,
    surfaceHeader,
  } = extractThemeColors(themeData, previewMode));

  async function toggle() {
    isOpen = !isOpen;
  }

  function handleModeChange(_: number, value: string) {
    previewMode = value.toLowerCase() as PreviewMode;
  }
</script>

<div
  class="relative w-full flex-none overflow-hidden flex flex-col bg-surface-background border-t"
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
      class="text-xs text-fg-primary rounded-sm hover:bg-surface-hover h-6 px-1.5 py-px flex items-center gap-1.5 transition-colors"
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
      class="preview-area overflow-y-auto px-4 py-4"
      class:dark={previewMode === "dark"}
      class:light={previewMode === "light"}
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
          {fgPrimary}
          {surfaceHeader}
        />
      </ThemeProvider>
    </div>
  {/if}
</div>

<style lang="postcss">
  .bar {
    @apply flex items-center px-3 h-7 w-full bg-surface-subtle border-b;
  }

  /* Force light mode colors when .light class is present, even if app is in dark mode */
  :global(.dark) .preview-area.light {
    color-scheme: light;
  }
</style>
