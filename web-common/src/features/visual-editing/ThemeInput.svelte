<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "../themes/color-config";
  import { useTheme } from "../themes/selectors";

  const defaultTheme: V1ThemeSpec = {
    primaryColorRaw: `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`,
    secondaryColorRaw: `hsl(${defaultSecondaryColors[500].split(" ").join(",")})`,
  };

  const fallbackCustomTheme: V1ThemeSpec = {
    primaryColorRaw: "hsl(180, 100%, 50%)",
    secondaryColorRaw: "lightgreen",
  };

  export let themeNames: string[];
  export let theme: string | V1ThemeSpec | undefined;
  export let onThemeChange: (themeName: string | undefined) => void;
  export let onColorChange: (primary: string, secondary: string) => void;
  export let small = false;

  let themeProxy: V1ThemeSpec =
    typeof theme === "string" || theme === undefined
      ? fallbackCustomTheme
      : theme;
  let presetProxy: string | undefined =
    typeof theme === "string" ? theme : undefined;

  $: ({ instanceId } = $runtime);

  $: themeQuery =
    typeof theme === "string" ? useTheme(instanceId, theme) : undefined;

  $: fetchedTheme = themeQuery && $themeQuery?.data?.theme?.spec;

  $: currentTheme =
    fetchedTheme ?? (theme === undefined ? defaultTheme : themeProxy);

  $: presetMode = theme === undefined || typeof theme === "string";
</script>

<div class="flex flex-col {small ? 'gap-y-2' : 'gap-y-1'}">
  <InputLabel
    label="Theme"
    {small}
    id="visual-explore-theme"
    hint="Colors may be adjusted for legibility"
  />

  <FieldSwitcher
    fields={["Presets", "Custom"]}
    selected={presetMode ? 0 : 1}
    onClick={(_, field) => {
      if (field === "Custom") {
        onColorChange(
          themeProxy.primaryColorRaw ?? "",
          themeProxy.secondaryColorRaw ?? "",
        );
        currentTheme = themeProxy;
      } else if (field === "Presets") {
        onThemeChange(presetProxy);
      }
    }}
  />
  <div class="gap-y-2 flex flex-col">
    {#if typeof theme === "string" || theme === undefined}
      <Select
        size={small ? "sm" : "lg"}
        fontSize={small ? 12 : 14}
        sameWidth
        onChange={(value) => {
          if (value === "Default") {
            onThemeChange(undefined);
            presetProxy = undefined;
          } else {
            onThemeChange(value);
            presetProxy = value;
          }
        }}
        value={theme ?? "Default"}
        options={["Default", ...themeNames].map((value) => ({
          value,
          label: value,
        }))}
        id="theme"
      />
    {/if}

    <ColorInput
      {small}
      stringColor={currentTheme.primaryColorRaw}
      label="Primary"
      disabled={presetMode}
      onChange={(color) => {
        onColorChange(color, themeProxy.secondaryColorRaw ?? "");
        themeProxy.primaryColorRaw = color;
      }}
    />
    <ColorInput
      {small}
      stringColor={currentTheme.secondaryColorRaw}
      label="Secondary"
      disabled={presetMode}
      onChange={(color) => {
        onColorChange(themeProxy.primaryColorRaw ?? "", color);
        themeProxy.secondaryColorRaw = color;
      }}
    />
  </div>
</div>
