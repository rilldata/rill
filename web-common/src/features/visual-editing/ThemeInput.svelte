<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { featureFlags } from "../feature-flags";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "../themes/color-config";
  import { useTheme } from "../themes/selectors";
  import { themeControl } from "../themes/theme-control";

  const DEFAULT_PRIMARY = `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`;
  const DEFAULT_SECONDARY = `hsl(${defaultSecondaryColors[500].split(" ").join(",")})`;
  const FALLBACK_PRIMARY = "hsl(180, 100%, 50%)";
  const FALLBACK_SECONDARY = "lightgreen";

  const { darkMode } = featureFlags;

  export let themeNames: string[];
  export let theme: string | V1ThemeSpec | undefined;
  export let small = false;
  export let onThemeChange: (themeName: string | undefined) => void;
  export let onColorChange: (
    primary: string,
    secondary: string,
    isDarkMode: boolean,
  ) => void;

  let customPrimary = "";
  let customSecondary = "";
  let lastPresetTheme: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: isPresetMode = theme === undefined || typeof theme === "string";
  $: embeddedTheme = typeof theme === "object" ? theme : undefined;

  $: lastPresetTheme =
    isPresetMode && typeof theme === "string" ? theme : lastPresetTheme;

  $: themeQuery =
    theme && typeof theme === "string"
      ? useTheme(instanceId, theme)
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
    : customPrimary || themePrimary || FALLBACK_PRIMARY;

  $: effectiveSecondary = isPresetMode
    ? themeSecondary || DEFAULT_SECONDARY
    : customSecondary || themeSecondary || FALLBACK_SECONDARY;

  $: currentSelectValue = isPresetMode
    ? typeof theme === "string"
      ? theme
      : "Default"
    : "Default";

  function handleModeSwitch(mode: string) {
    if (mode === "Custom") {
      if (!customPrimary) {
        customPrimary = themePrimary || FALLBACK_PRIMARY;
      }
      if (!customSecondary) {
        customSecondary = themeSecondary || FALLBACK_SECONDARY;
      }
      // Pass the current theme mode (light/dark)
      onColorChange(customPrimary, customSecondary, isDarkMode);
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
      customPrimary = color;
      // Pass the current theme mode (light/dark)
      onColorChange(customPrimary, effectiveSecondary, isDarkMode);
    } else {
      customSecondary = color;
      // Pass the current theme mode (light/dark)
      onColorChange(effectivePrimary, customSecondary, isDarkMode);
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
        options={["Default", ...themeNames].map((value) => ({
          value,
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
      allowLightnessControl={$darkMode}
      onChange={(color) => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        handleColorChange(color, true);
      }}
    />

    <ColorInput
      {small}
      stringColor={effectiveSecondary}
      label="Secondary"
      labelFirst
      disabled={isPresetMode}
      allowLightnessControl={$darkMode}
      onChange={(color) => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument
        handleColorChange(color, false);
      }}
    />
  </div>
</div>
