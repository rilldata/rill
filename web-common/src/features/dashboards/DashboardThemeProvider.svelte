<script lang="ts">
  import { page } from "$app/stores";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { setTheme } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const metricsView = useMetricsView(getStateManagers());
  $: themeFromUrl = $page.url.searchParams.get("theme");

  let theme: ReturnType<typeof useTheme>;
  $: themeName = themeFromUrl ?? $metricsView.data?.defaultTheme;
  $: if (themeName) theme = useTheme($runtime.instanceId, themeName);

  $: if ($theme?.data?.theme) {
    setTheme($theme.data.theme);
  }
  onMount(() => {
    // Handle the case where we have data in cache but the dashboard is not mounted yet
    if ($theme?.data?.theme) setTheme($theme.data.theme);
  });
</script>

<slot />
