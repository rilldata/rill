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

  const DEFAULT_PRIMARY = `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`;
  const DEFAULT_SECONDARY = `hsl(${defaultSecondaryColors[500].split(" ").join(",")})`;
  const FALLBACK_PRIMARY = "hsl(180, 100%, 50%)";
  const FALLBACK_SECONDARY = "lightgreen";

  const { darkMode } = featureFlags;

  export let themeNames: string[];
  export let theme: string | V1ThemeSpec | undefined;
  export let small = false;
  export let onThemeChange: (themeName: string | undefined) => void;
  export let onColorChange: (primary: string, secondary: string) => void;

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

  $: fetchedTheme = $themeQuery?.data?.theme?.spec as V1ThemeSpec | undefined;

  $: currentThemeSpec = embeddedTheme || fetchedTheme;

  $: themeColors = $darkMode
    ? (currentThemeSpec?.dark as
        | { primary?: string; secondary?: string }
        | undefined)
    : (currentThemeSpec?.light as
        | { primary?: string; secondary?: string }
        | undefined);

  $: primaryFromTheme =
    themeColors?.primary || currentThemeSpec?.primaryColorRaw;
  $: secondaryFromTheme =
    themeColors?.secondary || currentThemeSpec?.secondaryColorRaw;

  $: effectivePrimary = isPresetMode
    ? primaryFromTheme || DEFAULT_PRIMARY
    : customPrimary || primaryFromTheme || FALLBACK_PRIMARY;

  $: effectiveSecondary = isPresetMode
    ? secondaryFromTheme || DEFAULT_SECONDARY
    : customSecondary || secondaryFromTheme || FALLBACK_SECONDARY;

  $: currentSelectValue = isPresetMode
    ? typeof theme === "string"
      ? theme
      : "Default"
    : "Default";

  function handleModeSwitch(mode: string) {
    if (mode === "Custom") {
      if (!customPrimary) {
        customPrimary = primaryFromTheme || FALLBACK_PRIMARY;
      }
      if (!customSecondary) {
        customSecondary = secondaryFromTheme || FALLBACK_SECONDARY;
      }
      onColorChange(customPrimary, customSecondary);
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
      onColorChange(customPrimary, effectiveSecondary);
    } else {
      customSecondary = color;
      onColorChange(effectivePrimary, customSecondary);
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
      onChange={(color) => handleColorChange(color, true)}
    />

    <ColorInput
      {small}
      stringColor={effectiveSecondary}
      label="Secondary"
      labelFirst
      disabled={isPresetMode}
      allowLightnessControl={$darkMode}
      onChange={(color) => handleColorChange(color, false)}
    />
  </div>
</div>
