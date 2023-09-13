<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { useProjectRuntime } from "../../../components/projects/selectors";

  $: projRuntime = useProjectRuntime(
    $page.params.organization,
    $page.params.project
  );

  $: isRuntimeHibernating = $projRuntime.isSuccess && !$projRuntime.data;

  $: if (isRuntimeHibernating) {
    // Redirect any nested routes (notably dashboards) to the project page
    goto(`/${$page.params.organization}/${$page.params.project}`);
  }
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
