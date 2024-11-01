<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { setTheme } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const { validSpecStore } = getStateManagers();
  $: themeFromUrl = $page.url.searchParams.get("theme");

  let theme: ReturnType<typeof useTheme>;
  $: themeName = themeFromUrl ?? $validSpecStore.data?.explore?.theme;
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
