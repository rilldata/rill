<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { useTheme } from "../themes/selectors";

  export let themes: string[];
  export let selectedTheme: string;
</script>

<div class="flex flex-col gap-y-1">
  <InputLabel label="Theme" id="visual-explore-theme" />
  <Select
    fontSize={14}
    sameWidth
    onChange={async (value) => {
      if (value === "Custom") {
        await updateProperties({
          theme: {
            colors: {
              primary: "hsl(13, 98%, 54%)",
              secondary: "lightgreen",
            },
          },
        });
        return;
      } else if (value === "Default") {
        await updateProperties({}, ["theme"]);
      } else {
        await updateProperties({ theme: value });
      }
    }}
    value={!rawTheme
      ? "Default"
      : typeof rawTheme === "string"
        ? rawTheme
        : rawTheme instanceof YAMLMap
          ? "Custom"
          : undefined}
    options={["Default", ...themeNames, "Custom"].map((value) => ({
      value,
      label: value,
    }))}
    id="theme"
  />

  <!-- {#await theme then what} -->
  <div class="gap-y-2 flex flex-col">
    <ColorInput
      stringColor={what.primary}
      label="Primary"
      disabled={!what.custom}
      onChange={async (color) => {
        console.log("update");
        await updateProperties({
          theme: {
            colors: {
              primary: color,
              secondary: what.secondary,
            },
          },
        });
      }}
    />
    <ColorInput
      stringColor={what.secondary}
      label="Secondary"
      disabled={!what.custom}
      onChange={async (color) => {
        await updateProperties({
          theme: {
            colors: {
              primary: what.primary,
              secondary: color,
            },
          },
        });
      }}
    />
  </div>
  <!-- {/await} -->
</div>
