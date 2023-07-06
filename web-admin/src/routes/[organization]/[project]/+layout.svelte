<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { useProjectRuntime } from "../../../components/projects/selectors";

  $: projRuntime = useProjectRuntime(
    $page.params.organization,
    $page.params.project
  );
</script>

{#if $projRuntime.data}
  <RuntimeProvider
    host={$projRuntime.data.host}
    instanceId={$projRuntime.data.instanceId}
    jwt={$projRuntime.data?.jwt}
  >
    <slot />
  </RuntimeProvider>
{:else}
  <slot />
{/if}
