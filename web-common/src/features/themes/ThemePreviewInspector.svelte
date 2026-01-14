<script lang="ts">
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { Theme } from "./theme";
  import PreviewComponents from "./PreviewComponents.svelte";

  export let filePath: string;

  let previewMode: "light" | "dark" = "light";

  // Get file content and parse YAML to create Theme object
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ editorContent } = fileArtifact);

  $: parsedDocument = $editorContent ? parseDocument($editorContent) : null;
  $: themeData = (parsedDocument?.toJSON() || {}) as V1ThemeSpec;

  // Create Theme instance - this processes the YAML and generates CSS
  $: theme = themeData ? new Theme(themeData) : new Theme(undefined);

  // Extract colors directly from the parsed YAML (not from Theme class)
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
  $: backgroundColor =
    currentModeYaml["background"] ||
    (previewMode === "light" ? "#f9fafb" : "#111827");
  $: cardColor =
    currentModeYaml["card"] ||
    (previewMode === "light" ? "#ffffff" : "#374151");

  function handleModeChange(_: number, value: string) {
    previewMode = value.toLowerCase() as "light" | "dark";
  }
</script>

<Inspector {filePath}>
  <div class="preview-inspector">
    <div class="preview-header">
      <h3 class="preview-title">Preview</h3>
      <FieldSwitcher
        small
        fields={["Light", "Dark"]}
        selected={previewMode === "light" ? 0 : 1}
        onClick={handleModeChange}
      />
    </div>

    <div
      class="preview-content"
      class:dark={previewMode === "dark"}
      style="background-color: {backgroundColor};"
    >
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
