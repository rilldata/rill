<script lang="ts">
  import { page } from "$app/stores";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { setTheme } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  const { canvasEntity } = getCanvasStateManagers();
  const { canvasSpec } = canvasEntity.spec;
  $: themeFromUrl = $page.url.searchParams.get("theme");

  let theme: ReturnType<typeof useTheme>;
  $: themeName = themeFromUrl ?? $canvasSpec?.theme;
  $: if (themeName) theme = useTheme($runtime.instanceId, themeName);

  $: setTheme($theme?.data?.theme?.spec ?? $canvasSpec?.embeddedTheme);

  onMount(() => {
    // Handle the case where we have data in cache but the dashboard is not mounted yet
    setTheme($theme?.data?.theme?.spec ?? $canvasSpec?.embeddedTheme);
  });
</script>

<slot />
