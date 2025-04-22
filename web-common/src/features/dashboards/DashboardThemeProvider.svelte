<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { updateThemeVariables } from "@rilldata/web-common/features/themes/actions";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const { validSpecStore } = getStateManagers();

  let theme: ReturnType<typeof useTheme>;

  $: themeFromUrl = $page.url.searchParams.get("theme");

  $: ({ instanceId } = $runtime);
  $: themeName = themeFromUrl ?? $validSpecStore.data?.explore?.theme;
  $: if (themeName) theme = useTheme(instanceId, themeName);

  $: updateThemeVariables(
    $theme?.data?.theme?.spec ?? $validSpecStore?.data?.explore?.embeddedTheme,
  );
</script>

<slot />
