<script lang="ts">
  import { goto } from "$app/navigation";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import {
    themeControl,
    type ThemeMode,
  } from "@rilldata/web-common/features/themes/theme-control.ts";

  const { preference } = themeControl;
  $: selectedPreference = $preference;

  const ThemeOptions: {
    label: string;
    value: ThemeMode;
    image: string;
  }[] = [
    { label: "Light", value: "light", image: "/img/theme/light-mode.svg" },
    { label: "Dark", value: "dark", image: "/img/theme/dark-mode.svg" },
    {
      label: "System",
      value: "system",
      image: "/img/theme/system-mode.svg",
    },
  ];

  function handleThemeChange(theme: ThemeMode) {
    document.documentElement.classList.add("theme-transitioning");
    themeControl.set[theme]();
    setTimeout(
      () => document.documentElement.classList.remove("theme-transitioning"),
      300,
    );
  }

  function handleContinue() {
    themeControl.set[selectedPreference](); // Force selection so that localStorage is updated.
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
        class="flex flex-col gap-4"
        onclick={() => handleThemeChange(themeOption.value)}
        aria-label="Select {themeOption.label} theme"
      >
        <div
          class="border rounded-md transition-transform duration-200 hover:scale-110"
          class:shadow-lg={isSelected}
          class:border-ring-focus={isSelected}
        >
          <img src={themeOption.image} alt="{themeOption.value} image" />
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

<style>
  /* We are intentionally not changing `color`. It slows down the transition quite a bit. */
  :global(.theme-transitioning),
  :global(.theme-transitioning *) {
    transition:
      background-color 300ms ease,
      background-image 300ms ease,
      border-color 300ms ease,
      fill 300ms ease !important;
  }
</style>
