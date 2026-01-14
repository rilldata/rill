<script lang="ts">
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";

  export let fileArtifact: FileArtifact;
  export let filePath: string;
  export let errors: LineStatus[];

  let mode: "light" | "dark" = "light";

  function handleModeChange(_: number, value: string) {
    mode = value.toLowerCase() as "light" | "dark";
  }

  // Get file content and parse YAML directly
  $: fileArtifactFromStore = fileArtifacts.getFileArtifact(filePath);
  $: ({ editorContent, updateEditorContent } = fileArtifactFromStore);

  // Parse YAML document for editing
  $: parsedDocument = $editorContent ? parseDocument($editorContent) : null;
  $: themeData = parsedDocument?.toJSON() || {};

  $: lightTheme = themeData?.light || {};
  $: darkTheme = themeData?.dark || {};
  $: currentTheme = mode === "light" ? lightTheme : darkTheme;

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

    // Get or create the mode section (light/dark)
    let modeSection = parsedDocument.get(mode) as any;
    if (!modeSection) {
      parsedDocument.set(mode, {});
      modeSection = parsedDocument.get(mode) as any;
    }

    // Set the color value
    modeSection.set(colorKey, value);

    // Update the editor content
    updateEditorContent(parsedDocument.toString(), false, true);
  }

  // Core colors to edit
  const editableColors = [
    { key: "primary", label: "Primary" },
    { key: "secondary", label: "Secondary" },
    { key: "background", label: "Background" },
    { key: "surface", label: "Surface" },
    { key: "card", label: "Card" },
  ];

  const paletteInfo = [
    {
      name: "Diverging",
      count: 11,
      description: "For data with a midpoint (1-11)",
      variables: Array.from(
        { length: 11 },
        (_, i) => `color-diverging-${i + 1}`,
      ),
    },
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
  ];
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex items-center gap-x-4 pb-4 border-b">
      <h3 class="text-sm font-semibold text-gray-900">Theme Mode</h3>
      <FieldSwitcher
        small
        fields={["Light", "Dark"]}
        selected={mode === "light" ? 0 : 1}
        onClick={handleModeChange}
      />
    </div>

    <div class="content-scroll">
      <!-- Theme Colors Section -->
      <section class="section">
        <h3 class="section-title">Theme Colors</h3>
        <div class="palette-colors">
          {#each editableColors as { key, label }}
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
    @apply text-sm font-semibold text-gray-900;
  }

  .palette-section {
    @apply mt-4 space-y-3;
  }

  .palette-header {
    @apply flex items-center gap-x-4;
  }

  .palette-title {
    @apply text-sm font-semibold text-gray-800;
  }

  .palette-description {
    @apply text-xs text-gray-600;
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
    @apply text-xs font-semibold text-gray-700;
  }
</style>
