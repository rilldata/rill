<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createAdminServiceGetDeploymentCredentials } from "../../../client";
  import { viewAsUserStore } from "../../../components/authentication/viewAsUserStore";
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

  // if viewAs is set (which only admins can configure), we need to update the runtime with the new jwt
  $: deploymentCredsQuery = createAdminServiceGetDeploymentCredentials(
    $page.params.organization,
    $page.params.project,
    {
      userId: $viewAsUserStore?.id,
    },
    {
      query: {
        enabled: $viewAsUserStore?.id !== undefined,
      },
    }
  );
</script>

{#if $projRuntime.data}
  <RuntimeProvider
    host={$projRuntime.data.host}
    instanceId={$projRuntime.data.instanceId}
    jwt={$deploymentCredsQuery.data?.jwt || $projRuntime.data?.jwt}
  >
    <slot />
  </RuntimeProvider>
{:else}
  <slot />
{/if}
