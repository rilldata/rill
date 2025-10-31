<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { updateThemeVariables } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const { validSpecStore } = getStateManagers();

  let theme: ReturnType<typeof useTheme> | undefined;
  let themeBoundary: HTMLElement | null = null;

  $: themeFromUrl = $page.url.searchParams.get("theme");

  $: ({ instanceId } = $runtime);
  $: themeName = themeFromUrl ?? $validSpecStore.data?.explore?.theme;

  // Always update theme, even when themeName is undefined/null to reset
  $: {
    if (themeName) {
      theme = useTheme(instanceId, themeName);
    } else {
      theme = undefined;
    }
  }

  $: {
    const themeSpec = theme
      ? ($theme?.data?.theme?.spec ??
        $validSpecStore?.data?.explore?.embeddedTheme)
      : undefined;

    if (themeBoundary) {
      $themeControl;
      updateThemeVariables(themeSpec, themeBoundary);
    }
  }
</script>

<div class="dashboard-theme-boundary" bind:this={themeBoundary}>
  <slot />
</div>

<style>
  .dashboard-theme-boundary {
    /* Create a scoped boundary for theme variables while maintaining layout */
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }
</style>
