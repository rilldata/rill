<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { updateThemeVariables } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
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

  // Update theme variables, explicitly passing undefined when no theme to reset colors
  // Scope to the dashboard boundary element to avoid affecting the surrounding chrome
  $: {
    const themeSpec = theme
      ? ($theme?.data?.theme?.spec ??
        $validSpecStore?.data?.explore?.embeddedTheme)
      : undefined;

    // Only apply theme once we have the boundary element reference
    if (themeBoundary) {
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
