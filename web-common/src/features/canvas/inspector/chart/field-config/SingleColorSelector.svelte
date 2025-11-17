<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { colorToVariableReference } from "@rilldata/web-common/features/components/charts/util";
  import {
    primary,
    secondary,
  } from "@rilldata/web-common/features/themes/colors";
  import { type Color } from "chroma-js";
  export let markConfig: string;
  export let onChange: (newColor: string) => void;
  export let small = false;
  export let theme: {
    primary?: Color;
    secondary?: Color;
  };

  $: effectiveTheme = {
    primary: theme.primary || primary["500"],
    secondary: theme.secondary || secondary["500"],
  };

  $: isPresetMode = markConfig === "primary" || markConfig === "secondary";

  $: currentPreset = isPresetMode ? markConfig : "primary";

  $: displayColor = (() => {
    switch (markConfig) {
      case "primary":
        return effectiveTheme.primary.css("hsl");
      case "secondary":
        return effectiveTheme.secondary.css("hsl");
      default:
        return markConfig;
    }
  })();

  $: colorLabel = (() => {
    switch (markConfig) {
      case "primary":
        return "Primary";
      case "secondary":
        return "Secondary";
      default:
        return markConfig;
    }
  })();

  let customColor = "";

  function handleModeSwitch(mode: string) {
    if (mode === "Custom") {
      if (!customColor) {
        customColor = displayColor;
      }
      onChange(customColor);
    } else {
      onChange(currentPreset);
    }
  }

  function handlePresetSelection(value: string) {
    onChange(value);
  }

  function handleColorChange(color: string) {
    customColor = color;
    // Convert color back to CSS variable reference if it matches a palette color
    const colorToSave = colorToVariableReference(color);
    onChange(colorToSave);
  }
</script>

<div class="flex flex-col {small ? 'gap-y-2' : 'gap-y-1'}">
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
        onChange={handlePresetSelection}
        value={currentPreset}
        options={[
          { value: "primary", label: "Primary" },
          { value: "secondary", label: "Secondary" },
        ]}
        id="color-preset-select"
      />
    {/if}

    <ColorInput
      {small}
      stringColor={displayColor}
      label={isPresetMode ? colorLabel : ""}
      disabled={isPresetMode}
      allowLightnessControl
      onChange={handleColorChange}
    />
  </div>
</div>
