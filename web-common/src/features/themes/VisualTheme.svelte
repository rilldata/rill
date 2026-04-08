<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";
  import { themePreviewMode, updateThemeColor } from "./theme-preview-utils";

  interface Props {
    filePath: string;
  }

  let { filePath }: Props = $props();

  function handleModeChange(_: number, value: string) {
    themePreviewMode.set(value.toLowerCase() as "light" | "dark");
  }

  // Get file content and parse YAML directly
  let fileArtifactFromStore = $derived(fileArtifacts.getFileArtifact(filePath));
  let editorContent = $derived(fileArtifactFromStore.editorContent);
  let updateEditorContent = $derived(fileArtifactFromStore.updateEditorContent);

  // Parse YAML document for editing
  let parsedDocument = $derived(
    $editorContent ? parseDocument($editorContent) : null,
  );
  let themeData = $derived(parsedDocument?.toJSON() || {});

  let lightTheme = $derived(themeData?.light || {});
  let darkTheme = $derived(themeData?.dark || {});
  let currentTheme = $derived(
    $themePreviewMode === "light" ? lightTheme : darkTheme,
  );

  // Combine primary/secondary with other color variables
  let currentColors: Record<string, string> = $derived.by(() => {
    return {
      primary: currentTheme?.primary || "",
      secondary: currentTheme?.secondary || "",
      ...Object.fromEntries(
        Object.entries(currentTheme).filter(
          ([key]) =>
            key.startsWith("color-") || !["primary", "secondary"].includes(key),
        ),
      ),
    };
  });

  // Check if the theme has any color values defined
  let hasAnyColors = $derived(
    Object.values(lightTheme).some((v) => !!v) ||
      Object.values(darkTheme).some((v) => !!v),
  );

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

  const textColors = [{ key: "fg-primary", label: "Text Primary" }];

  const kpiColors = [
    { key: "kpi-positive", label: "Positive" },
    { key: "kpi-negative", label: "Negative" },
  ];

  const paletteInfo = [
    {
      name: "Qualitative",
      description: "For categorical data (1-24)",
      variables: Array.from(
        { length: 24 },
        (_, i) => `color-qualitative-${i + 1}`,
      ),
    },
    {
      name: "Sequential",
      description: "For continuous data (1-9)",
      variables: Array.from(
        { length: 9 },
        (_, i) => `color-sequential-${i + 1}`,
      ),
    },
    {
      name: "Diverging",
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
    <div class="flex flex-col gap-y-3 pb-4 border-b">
      <div class="flex items-center gap-x-4">
        <h3 class="text-sm font-semibold text-fg-primary">Theme Mode</h3>
        <FieldSwitcher
          small
          fields={["Light", "Dark"]}
          selected={$themePreviewMode === "light" ? 0 : 1}
          onClick={handleModeChange}
        />
      </div>
      <!-- Live swatch preview of core colors for the active mode -->
      {#if hasAnyColors}
        <div class="mode-preview">
          <div
            class="mode-preview-bg"
            style:background-color={currentColors["surface-background"] ||
              "#fff"}
          >
            <!-- Header bar -->
            {#if currentColors["surface-subtle"]}
              <div
                class="mode-preview-header"
                style:background-color={currentColors["surface-subtle"]}
              >
                {#if currentColors["fg-primary"]}
                  <span
                    class="text-[10px] font-semibold"
                    style:color={currentColors["fg-primary"]}
                    >Primary Text Color</span
                  >
                {/if}
              </div>
            {/if}
            <!-- Component cards with color dots -->
            <div class="mode-preview-cards">
              {#each ["primary", "secondary"] as key (key)}
                <div
                  class="mode-preview-card"
                  style:background-color={currentColors["surface-card"] ||
                    "#fff"}
                >
                  {#if currentColors[key]}
                    <div
                      class="mode-preview-dot"
                      style:background-color={currentColors[key]}
                    ></div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        </div>
      {/if}
    </div>

    <div class="content-scroll">
      {#if !hasAnyColors}
        <div class="empty-state">
          <p class="text-sm text-fg-secondary">
            No colors defined yet. Choose a primary color to get started.
          </p>
        </div>
      {/if}

      <!-- Core Colors Section -->
      <section class="section">
        <h3 class="section-title">Core Colors</h3>
        <div class="palette-colors">
          {#each coreColors as { key, label } (key)}
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

      <!-- Text Section -->
      <section class="section">
        <h3 class="section-title">Text</h3>
        <div class="palette-colors">
          {#each textColors as { key, label } (key)}
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

      <!-- KPI Section -->
      <section class="section">
        <h3 class="section-title">KPI</h3>
        <div class="palette-colors">
          {#each kpiColors as { key, label } (key)}
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

        {#each paletteInfo as palette (palette.name)}
          <div class="palette-section">
            <div class="palette-header">
              <h4 class="palette-title">{palette.name}</h4>
              <p class="palette-description">{palette.description}</p>
            </div>
            <!-- Color strip preview -->
            <div class="palette-strip">
              {#each palette.variables as variable (variable)}
                {@const color = currentColors[variable]}
                <div
                  class="palette-strip-cell"
                  style:background-color={color || "#e5e7eb"}
                ></div>
              {/each}
            </div>
            <div class="palette-colors">
              {#each palette.variables as variable (variable)}
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
    @apply size-full overflow-hidden flex flex-col;
  }

  .main-area {
    @apply flex flex-col gap-y-4 flex-1 p-4 bg-surface-background border overflow-hidden;
  }

  .content-scroll {
    @apply flex-1 overflow-y-auto space-y-6;
  }

  .mode-preview {
    @apply w-full;
  }

  .mode-preview-bg {
    @apply flex flex-col rounded border overflow-hidden;
  }

  .mode-preview-header {
    @apply flex items-center gap-2 px-2 py-1.5;
  }

  .mode-preview-dot {
    @apply size-3 rounded-full border border-black/10;
  }

  .mode-preview-cards {
    @apply flex gap-2 p-2;
  }

  .mode-preview-card {
    @apply flex-1 h-8 rounded border border-black/10 flex items-center pl-2;
  }

  .empty-state {
    @apply rounded border border-dashed border-gray-300 p-4 text-center;
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

  .palette-strip {
    @apply flex rounded overflow-hidden h-6;
  }

  .palette-strip-cell {
    @apply flex-1 min-w-0;
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
