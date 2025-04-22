<script lang="ts">
  import { page } from "$app/stores";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let canvasName: string;

  let theme: ReturnType<typeof useTheme>;

  $: ({
    canvasEntity: { setTheme, spec },
  } = getCanvasStore(canvasName));

  $: ({ canvasSpec: cs } = spec);
  $: ({ instanceId } = $runtime);

  $: themeFromUrl = $page.url.searchParams.get("theme");

  $: canvasSpec = $cs;

  $: themeName = themeFromUrl ?? canvasSpec?.theme;
  $: if (themeName) theme = useTheme(instanceId, themeName);

  $: setTheme($theme?.data?.theme?.spec ?? canvasSpec?.embeddedTheme);
</script>

<slot />
