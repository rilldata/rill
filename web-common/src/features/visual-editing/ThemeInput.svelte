<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Pencil } from "lucide-svelte";
  import { useTheme } from "../themes/selectors";
  import { themeControl } from "../themes/theme-control";
  import { themeEditorStore } from "./theme-editor-store";
  import ThemePropertySections from "./ThemePropertySections.svelte";

  const runtimeClient = useRuntimeClient();

  export let themeNames: string[];
  export let theme: string | V1ThemeSpec | undefined;
  export let projectDefaultTheme: string | undefined = undefined;
  export let small = false;
  export let onThemeChange: (themeName: string | undefined) => void;
  export let onInlineThemeChange: (spec: V1ThemeSpec) => void;

  $: isPresetMode = theme === undefined || typeof theme === "string";
  $: currentThemeName = typeof theme === "string" ? theme : undefined;

  $: themeQuery = currentThemeName
    ? useTheme(runtimeClient, currentThemeName)
    : !currentThemeName && projectDefaultTheme
      ? useTheme(runtimeClient, projectDefaultTheme)
      : undefined;

  $: fetchedSpec = $themeQuery?.data?.theme?.spec;
  $: themeResource = $themeQuery?.data;

  $: isDarkMode = $themeControl === "dark";

  $: editorState = $themeEditorStore;
  $: editing = editorState.editing;

  $: inlineSpec =
    theme && typeof theme === "object" ? (theme as V1ThemeSpec) : undefined;

  $: displayValues =
    editing && editorState.mode === "custom"
      ? themeEditorStore.getValues(isDarkMode)
      : getSpecValues(isPresetMode ? fetchedSpec : inlineSpec, isDarkMode);

  $: currentSelectValue = typeof theme === "string" ? theme : "Default";

  $: hasTheme = !!currentThemeName || !!projectDefaultTheme || !isPresetMode;

  $: themeFilePath = themeResource?.meta?.filePaths?.[0];

  function getSpecValues(
    spec: V1ThemeSpec | undefined,
    dark: boolean,
  ): Record<string, string> {
    if (!spec) return {};
    const modeColors = dark ? spec.dark : spec.light;
    if (!modeColors) return {};
    const result: Record<string, string> = {};
    if (modeColors.primary) result.primary = modeColors.primary;
    if (modeColors.secondary) result.secondary = modeColors.secondary;
    if (modeColors.variables) Object.assign(result, modeColors.variables);
    return result;
  }

  function handleThemeSelection(value: string) {
    themeEditorStore.exitEditing();
    if (value === "Default") {
      onThemeChange(undefined);
    } else {
      onThemeChange(value);
    }
  }

  function handleModeSwitch(_: number, field: string) {
    if (field === "Custom") {
      enterCustomMode();
    } else {
      themeEditorStore.exitEditing();
      onThemeChange(currentThemeName);
    }
  }

  function enterCustomMode() {
    const baseSpec: V1ThemeSpec = fetchedSpec ?? {
      light: { primary: "hsl(180, 100%, 50%)", secondary: "lightgreen" },
      dark: { primary: "hsl(180, 100%, 50%)", secondary: "lightgreen" },
    };
    onInlineThemeChange(baseSpec);
    themeEditorStore.startCustom(baseSpec);
  }

  let saveTimer: ReturnType<typeof setTimeout> | undefined;

  // Only the Custom view is editable; Presets is view-only and routes the
  // user to the YAML file via `Edit theme file`.
  function handlePropertyChange(key: string, value: string) {
    if (isPresetMode) return;
    if (!editing && inlineSpec) {
      themeEditorStore.startCustom(inlineSpec);
    }
    themeEditorStore.updateProperty(key, value, isDarkMode);
    debouncedSave();
  }

  function debouncedSave() {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      onInlineThemeChange(themeEditorStore.buildSpec());
      themeEditorStore.markSaved();
    }, 800);
  }

  function handleEditThemeFile() {
    if (!themeFilePath) return;
    void navigateToFile(themeFilePath);
  }
</script>

<div class="flex flex-col {small ? 'gap-y-2' : 'gap-y-1'}">
  <InputLabel label="Theme" {small} id="visual-explore-theme" />

  <FieldSwitcher
    {small}
    expand
    fields={["Presets", "Custom"]}
    selected={isPresetMode ? 0 : 1}
    onClick={(i, field) => {
      handleModeSwitch(i, field);
    }}
  />

  {#if isPresetMode}
    <Select
      size={small ? "sm" : "lg"}
      fontSize={small ? 12 : 14}
      sameWidth
      onChange={handleThemeSelection}
      value={currentSelectValue}
      options={[
        projectDefaultTheme ? `Default (${projectDefaultTheme})` : "Default",
        ...themeNames,
      ].map((value) => ({
        value: value.startsWith("Default") ? "Default" : value,
        label: value,
      }))}
      id="theme"
    />
  {/if}

  {#if hasTheme || editing}
    <div class="sections-container">
      <ThemePropertySections
        values={displayValues}
        onPropertyChange={handlePropertyChange}
        readonly={isPresetMode}
      />
    </div>

    {#if isPresetMode && themeFilePath}
      <button type="button" class="edit-file-btn" onclick={handleEditThemeFile}>
        <Pencil size="12" />
        <span>Edit theme file</span>
      </button>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .sections-container {
    @apply max-h-[60vh] overflow-y-auto;
  }

  .edit-file-btn {
    @apply mt-1 inline-flex items-center gap-x-1.5 self-start;
    @apply text-xs font-medium text-fg-secondary;
    @apply px-2 py-1 rounded border border-border bg-surface-base;
  }

  .edit-file-btn:hover {
    @apply bg-surface-hover text-fg-primary;
  }
</style>
