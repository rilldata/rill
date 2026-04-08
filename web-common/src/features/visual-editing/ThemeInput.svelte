<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useTheme } from "../themes/selectors";
  import { themeControl } from "../themes/theme-control";
  import { themeEditorStore } from "./theme-editor-store";
  import { buildThemeYaml } from "./theme-yaml-utils";
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

  $: displayValues = editing
    ? themeEditorStore.getValues(isDarkMode)
    : getSpecValues(isPresetMode ? fetchedSpec : inlineSpec, isDarkMode);

  $: currentSelectValue = typeof theme === "string" ? theme : "Default";

  $: hasTheme = !!currentThemeName || !!projectDefaultTheme || !isPresetMode;

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
      light: { primary: "#6366f1", secondary: "#8b5cf6" },
      dark: { primary: "#818cf8", secondary: "#a78bfa" },
    };
    onInlineThemeChange(baseSpec);
    themeEditorStore.startCustom(baseSpec);
  }

  let saveTimer: ReturnType<typeof setTimeout> | undefined;

  function handlePropertyChange(key: string, value: string) {
    if (!editing) {
      const spec = isPresetMode ? fetchedSpec : inlineSpec;
      if (spec) {
        if (isPresetMode) {
          const themeName = currentThemeName ?? projectDefaultTheme;
          themeEditorStore.startEditing(spec, themeName);
        } else {
          themeEditorStore.startCustom(spec);
        }
      }
    }
    themeEditorStore.updateProperty(key, value, isDarkMode);
    debouncedSave();
  }

  function debouncedSave() {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(autoSave, 800);
  }

  async function autoSave() {
    const spec = themeEditorStore.buildSpec();

    if (!isPresetMode || editorState.mode === "custom") {
      onInlineThemeChange(spec);
    } else {
      const yamlContent = buildThemeYaml(spec);
      const filePath = themeResource?.meta?.filePaths?.[0];
      if (!filePath) return;

      await runtimeServicePutFile(runtimeClient, {
        path: filePath,
        blob: yamlContent,
      });
    }

    themeEditorStore.markSaved();
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
      />
    </div>
  {/if}
</div>

<style lang="postcss">
  .sections-container {
    @apply max-h-[60vh] overflow-y-auto;
  }
</style>
