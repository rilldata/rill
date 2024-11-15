<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "../themes/color-config";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { useTheme } from "../themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const defaultTheme: V1ThemeSpec = {
    primaryColorRaw: `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`,
    secondaryColorRaw: `hsl(${defaultSecondaryColors[500].split(" ").join(",")})`,
  };

  export let themeName: string | "Custom" | "Default" | undefined;
  export let themeNames: string[];
  export let theme: V1ThemeSpec | undefined;
  export let onModeChange: (mode: string) => void;
  export let onColorChange: (primary: string, secondary: string) => void;

  $: ({ instanceId } = $runtime);

  $: themeQuery = useTheme(instanceId, themeName ?? "");

  $: fetchedTheme = $themeQuery?.data?.theme?.spec;

  $: theme = theme ?? fetchedTheme ?? defaultTheme;
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel
    label="Theme"
    id="visual-explore-theme"
    hint="Colors may be adjusted for legibility"
  />
  <div class="gap-y-2 flex flex-col">
    <Select
      fontSize={14}
      sameWidth
      onChange={onModeChange}
      value={themeName}
      options={["Default", ...themeNames, "Custom"].map((value) => ({
        value,
        label: value,
      }))}
      id="theme"
    />

    <ColorInput
      stringColor={theme.primaryColorRaw}
      label="Primary"
      disabled={themeName !== "Custom" && themeName !== "default"}
      onChange={(color) => {
        onColorChange(color, theme.secondaryColorRaw ?? "");
      }}
    />
    <ColorInput
      stringColor={theme.secondaryColorRaw}
      label="Secondary"
      disabled={themeName !== "Custom" && themeName !== "default"}
      onChange={(color) => {
        onColorChange(theme.primaryColorRaw ?? "", color);
      }}
    />
  </div>
</div>
