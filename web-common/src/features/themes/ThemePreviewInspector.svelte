<script lang="ts">
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import PreviewComponents from "./PreviewComponents.svelte";
  import {
    parseThemeFromYaml,
    extractThemeColors,
    themePreviewMode,
  } from "./theme-preview-utils";

  export let filePath: string;

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
  } = extractThemeColors(themeData, $themePreviewMode));

  function handleModeChange(_: number, value: string) {
    themePreviewMode.set(value.toLowerCase() as "light" | "dark");
  }
</script>

<Inspector {filePath}>
  <div class="preview-inspector">
    <div class="preview-header">
      <h3 class="preview-title">Preview</h3>
      <FieldSwitcher
        small
        fields={["Light", "Dark"]}
        selected={$themePreviewMode === "light" ? 0 : 1}
        onClick={handleModeChange}
      />
    </div>

    <div
      class="preview-content"
      class:dark={$themePreviewMode === "dark"}
      style="background-color: {backgroundColor};"
    >
      <ThemeProvider {theme}>
        <PreviewComponents
          {sequentialColors}
          {qualitativeColors}
          {divergingColors}
          {primaryColor}
          {cardColor}
          {fgPrimary}
        />
      </ThemeProvider>
    </div>
  </div>
</Inspector>

<style lang="postcss">
  .preview-inspector {
    @apply flex flex-col h-full;
  }

  .preview-header {
    @apply flex items-center justify-between p-3 border-b border-gray-200;
  }

  .preview-title {
    @apply text-sm font-semibold text-gray-900;
  }

  .preview-content {
    @apply flex-1 overflow-y-auto p-3;
  }
</style>
