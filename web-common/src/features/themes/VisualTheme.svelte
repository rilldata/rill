<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";
  import { themePreviewMode, updateThemeColor } from "./theme-preview-utils";

  export let filePath: string;

  function handleModeChange(_: number, value: string) {
    themePreviewMode.set(value.toLowerCase() as "light" | "dark");
  }

  // Get file content and parse YAML directly
  $: fileArtifactFromStore = fileArtifacts.getFileArtifact(filePath);
  $: ({ editorContent, updateEditorContent } = fileArtifactFromStore);

  // Parse YAML document for editing
  $: parsedDocument = $editorContent ? parseDocument($editorContent) : null;
  $: themeData = parsedDocument?.toJSON() || {};

  $: lightTheme = themeData?.light || {};
  $: darkTheme = themeData?.dark || {};
  $: currentTheme = $themePreviewMode === "light" ? lightTheme : darkTheme;

  // Combine primary/secondary with other color variables
  $: currentColors = {
    primary: currentTheme?.primary || "",
    secondary: currentTheme?.secondary || "",
    ...Object.fromEntries(
      Object.entries(currentTheme).filter(
        ([key]) =>
          key.startsWith("color-") || !["primary", "secondary"].includes(key),
      ),
    ),
  };

  function updateColor(colorKey: string, value: string) {
    if (!parsedDocument) return;

    const updatedYaml = updateThemeColor(
      parsedDocument,
      $themePreviewMode,
      colorKey,
      value,
    );
    updateEditorContent(updatedYaml, false, true);
  }

  // Core colors to edit
  const coreColors = [
    { key: "primary", label: "Primary" },
    { key: "secondary", label: "Secondary" },
    { key: "surface-background", label: "Surface Background" },
    { key: "surface-subtle", label: "Surface Header" },
    { key: "surface-card", label: "Component Background" },
  ];

  // Text/Foreground colors
  const textColors = [
    { key: "fg-primary", label: "Primary (Titles)" },
    // Keeping these commented out for now, but they can be easily added back if needed. These should be surfaced for true theme editing.
    // { key: "fg-secondary", label: "Secondary (Body)" },
    // { key: "fg-tertiary", label: "Tertiary (Captions)" },
    // { key: "fg-muted", label: "Muted (Hints)" },
    // { key: "fg-disabled", label: "Disabled" },
    // { key: "fg-inverse", label: "Inverse (On Dark)" },
    // { key: "fg-accent", label: "Accent (Links)" },
  ];

  const paletteInfo = [
    {
      name: "Qualitative",
      count: 24,
      description: "For categorical data (1-24)",
      variables: Array.from(
        { length: 24 },
        (_, i) => `color-qualitative-${i + 1}`,
      ),
    },
    {
      name: "Sequential",
      count: 9,
      description: "For continuous data (1-9)",
      variables: Array.from(
        { length: 9 },
        (_, i) => `color-sequential-${i + 1}`,
      ),
    },
    {
      name: "Diverging",
      count: 11,
      description: "For data with a midpoint (1-11)",
      variables: Array.from(
        { length: 11 },
        (_, i) => `color-diverging-${i + 1}`,
      ),
    },
  ];
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex items-center gap-x-4 pb-4 border-b">
      <h3 class="text-sm font-semibold text-fg-primary">Theme Mode</h3>
      <FieldSwitcher
        small
        fields={["Light", "Dark"]}
        selected={$themePreviewMode === "light" ? 0 : 1}
        onClick={handleModeChange}
      />
    </div>

    <div class="content-scroll">
      <!-- Core Colors Section -->
      <section class="section">
        <h3 class="section-title">Core Colors</h3>
        <div class="palette-colors">
          {#each coreColors as { key, label }}
            <div class="theme-color-item">
              <span class="theme-color-label">{label}</span>
              <ColorInput
                label=""
                stringColor={currentColors[key] || ""}
                onChange={(color) => updateColor(key, color)}
                allowLightnessControl={true}
                small={true}
              />
            </div>
          {/each}
        </div>
      </section>

      <!-- Text Colors Section -->
      <section class="section">
        <h3 class="section-title">Text Colors</h3>
        <div class="palette-colors">
          {#each textColors as { key, label }}
            <div class="theme-color-item">
              <span class="theme-color-label">{label}</span>
              <ColorInput
                label=""
                stringColor={currentColors[key] || ""}
                onChange={(color) => updateColor(key, color)}
                allowLightnessControl={true}
                small={true}
              />
            </div>
          {/each}
        </div>
      </section>

      <!-- Color Palettes Section -->
      <section class="section">
        <h3 class="section-title">Color Palettes</h3>

        {#each paletteInfo as palette}
          <div class="palette-section">
            <div class="palette-header">
              <h4 class="palette-title">{palette.name}</h4>
              <p class="palette-description">{palette.description}</p>
            </div>
            <div class="palette-colors">
              {#each palette.variables as variable}
                {@const color = currentColors[variable]}
                <div class="palette-color-item">
                  <ColorInput
                    label=""
                    stringColor={color || ""}
                    onChange={(newColor) => updateColor(variable, newColor)}
                    allowLightnessControl={true}
                    small={true}
                  />
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </section>
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply size-full overflow-hidden;
  }

  .main-area {
    @apply flex flex-col gap-y-4 flex-1 p-4 bg-surface overflow-hidden;
  }

  .content-scroll {
    @apply flex-1 overflow-y-auto space-y-6;
  }

  .section {
    @apply space-y-3;
  }

  .section-title {
    @apply text-sm font-semibold text-fg-primary;
  }

  .palette-section {
    @apply mt-4 space-y-3;
  }

  .palette-header {
    @apply flex items-center gap-x-4;
  }

  .palette-title {
    @apply text-sm font-semibold text-fg-primary;
  }

  .palette-description {
    @apply text-xs text-fg-secondary;
  }

  .palette-colors {
    @apply flex flex-wrap gap-2;
  }

  .palette-color-item {
    @apply flex items-center;
  }

  .theme-color-item {
    @apply flex flex-col items-start gap-1;
  }

  .theme-color-label {
    @apply text-xs font-semibold text-fg-secondary;
  }
</style>
