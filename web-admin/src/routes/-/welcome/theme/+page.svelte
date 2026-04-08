<script lang="ts">
  import { goto } from "$app/navigation";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import {
    themeControl,
    type ThemeMode,
  } from "@rilldata/web-common/features/themes/theme-control.ts";
  import LightModeIcon from "@rilldata/web-admin/features/welcome/icons/LightModeIcon.svelte";
  import DarkModeIcon from "@rilldata/web-admin/features/welcome/icons/DarkModeIcon.svelte";
  import SystemModeIcon from "@rilldata/web-admin/features/welcome/icons/SystemModeIcon.svelte";

  const { preference } = themeControl;
  $: selectedPreference = $preference;

  const ThemeOptions: {
    label: string;
    value: ThemeMode;
    icon: typeof LightModeIcon | typeof DarkModeIcon | typeof SystemModeIcon;
  }[] = [
    { label: "Light", value: "light", icon: LightModeIcon },
    { label: "Dark", value: "dark", icon: DarkModeIcon },
    { label: "System", value: "system", icon: SystemModeIcon },
  ];

  function handleThemeChange(theme: ThemeMode) {
    themeControl.set[theme]();
  }

  function handleContinue() {
    return goto("/-/welcome/organization");
  }
</script>

<div class="flex flex-col gap-4 justify-center">
  <RillLogoSquareNegative size="36px" />
  <div class="text-2xl font-extrabold text-fg-accent text-center">
    Pick your color mode
  </div>
  <div class="flex flex-row gap-8 pt-6 mx-auto">
    {#each ThemeOptions as themeOption (themeOption.value)}
      {@const isSelected = selectedPreference === themeOption.value}
      <button
        class="flex flex-col gap-2"
        onclick={() => handleThemeChange(themeOption.value)}
      >
        <div
          class="border rounded-md"
          class:shadow-lg={isSelected}
          class:border-ring-focus={isSelected}
        >
          <svelte:component this={themeOption.icon} />
        </div>
        <div class="text-sm font-semibold text-fg-primary">
          {themeOption.label}
        </div>
      </button>
    {/each}
  </div>
  <div class="mx-auto pt-12 pb-24">
    <Button type="primary" onClick={handleContinue} large>Continue</Button>
  </div>
</div>
