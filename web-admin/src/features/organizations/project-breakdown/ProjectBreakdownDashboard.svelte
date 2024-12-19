<script lang="ts">
  import ExploreEmbed from "@rilldata/web-admin/features/embeds/ExploreEmbed.svelte";
  import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: instanceId = $runtime.instanceId;
  $: explores = useValidExplores(instanceId);
  $: exploreName = $explores.data?.[0]?.meta?.name?.name;
</script>

{#if exploreName}
  <ExploreEmbed {instanceId} {exploreName} />
{/if}
