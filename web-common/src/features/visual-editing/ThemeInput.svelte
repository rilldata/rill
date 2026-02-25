<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "../themes/color-config";
  import { useTheme } from "../themes/selectors";
  import { themeControl } from "../themes/theme-control";

  const runtimeClient = useRuntimeClient();

  const DEFAULT_PRIMARY = `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`;
  const DEFAULT_SECONDARY = `hsl(${defaultSecondaryColors[500].split(" ").join(",")})`;
  const FALLBACK_PRIMARY = "hsl(180, 100%, 50%)";
  const FALLBACK_SECONDARY = "lightgreen";

  export let themeNames: string[];
  export let theme: string | V1ThemeSpec | undefined;
  export let projectDefaultTheme: string | undefined = undefined;
  export let small = false;
  export let onThemeChange: (themeName: string | undefined) => void;
  export let onColorChange: (
    primary: string,
    secondary: string,
    isDarkMode: boolean,
  ) => void;

  let lastPresetTheme: string | undefined = undefined;

  $: ({ instanceId } = runtimeClient);

  $: isPresetMode = theme === undefined || typeof theme === "string";
  $: embeddedTheme = typeof theme === "object" ? theme : undefined;

  $: lastPresetTheme =
    isPresetMode && typeof theme === "string" ? theme : lastPresetTheme;

  $: themeQuery =
    theme && typeof theme === "string"
      ? useTheme(instanceId, theme)
      : !theme && projectDefaultTheme
        ? useTheme(instanceId, projectDefaultTheme)
        : undefined;

  $: fetchedTheme = $themeQuery?.data?.theme?.spec;

  $: currentThemeSpec = embeddedTheme || fetchedTheme;

  // Determine if we're in dark mode
  $: isDarkMode = $themeControl === "dark";

  // Extract colors from the appropriate theme section (light or dark)
  $: themePrimary = isDarkMode
    ? currentThemeSpec?.dark?.primary || currentThemeSpec?.primaryColorRaw
    : currentThemeSpec?.light?.primary || currentThemeSpec?.primaryColorRaw;

  $: themeSecondary = isDarkMode
    ? currentThemeSpec?.dark?.secondary || currentThemeSpec?.secondaryColorRaw
    : currentThemeSpec?.light?.secondary || currentThemeSpec?.secondaryColorRaw;

  $: effectivePrimary = isPresetMode
    ? themePrimary || DEFAULT_PRIMARY
    : themePrimary || FALLBACK_PRIMARY;

  $: effectiveSecondary = isPresetMode
    ? themeSecondary || DEFAULT_SECONDARY
    : themeSecondary || FALLBACK_SECONDARY;

  $: currentSelectValue = isPresetMode
    ? typeof theme === "string"
      ? theme
      : "Default"
    : "Default";

  function handleModeSwitch(mode: string) {
    if (mode === "Custom") {
      // Pass the current theme mode (light/dark)
      onColorChange(
        themePrimary || FALLBACK_PRIMARY,
        themeSecondary || FALLBACK_SECONDARY,
        isDarkMode,
      );
    } else {
      onThemeChange(lastPresetTheme);
    }
  }

  function handleThemeSelection(value: string) {
    if (value === "Default") {
      lastPresetTheme = undefined;
      onThemeChange(undefined);
    } else {
      lastPresetTheme = value;
      onThemeChange(value);
    }
  }

  function handleColorChange(color: string, isPrimary: boolean) {
    if (isPrimary) {
      // Pass the current theme mode (light/dark)
      onColorChange(color, effectiveSecondary, isDarkMode);
    } else {
      // Pass the current theme mode (light/dark)
      onColorChange(effectivePrimary, color, isDarkMode);
    }
  }
</script>

<div class="flex flex-col {small ? 'gap-y-2' : 'gap-y-1'}">
  <InputLabel
    label="Theme"
    {small}
    id="visual-explore-theme"
    hint="Colors may be adjusted for legibility"
  />

  <FieldSwitcher
    {small}
    expand
    fields={["Presets", "Custom"]}
    selected={isPresetMode ? 0 : 1}
    onClick={(_, field) => {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
      handleModeSwitch(field);
    }}
  />

  <div class="gap-y-2 flex flex-col">
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

    <ColorInput
      {small}
      stringColor={effectivePrimary}
      label="Primary"
      labelFirst
      disabled={isPresetMode}
      allowLightnessControl
      onChange={(color) => {
        handleColorChange(color, true);
      }}
    />

    <ColorInput
      {small}
      stringColor={effectiveSecondary}
      label="Secondary"
      labelFirst
      disabled={isPresetMode}
      allowLightnessControl
      onChange={(color) => {
        handleColorChange(color, false);
      }}
    />
  </div>
</div>
