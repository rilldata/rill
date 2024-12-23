<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import ExploreEmbed from "@rilldata/web-admin/features/embeds/ExploreEmbed.svelte";
  import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: instanceId = $runtime.instanceId;
  $: explores = useValidExplores(instanceId);
  $: exploreName = $explores.data?.find(
    (e) => e.meta?.name?.name === "db_size_explore",
  )?.meta?.name?.name;

  beforeNavigate(({ from, to }) => {
    if (!from || !to || from.url.pathname === to.url.pathname) {
      // routing to the same path but probably different url params
      return;
    }

    metricsExplorerStore.clearAllFilters(exploreName);
  });
</script>

{#if exploreName}
  <ExploreEmbed {instanceId} {exploreName} />
{/if}
