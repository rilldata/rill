<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createAdminServiceGetProject } from "../../../client";

  $: projRuntime = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project,
    {
      query: {
        // Proactively refetch the JWT because it's only valid for 1 hour
        refetchInterval: 1000 * 60 * 30, // 30 minutes
        select: (data) => {
          return {
            // Hack: in development, the runtime host is actually on port 8081
            host: data.productionDeployment.runtimeHost.replace(
              "localhost:9091",
              "localhost:8081"
            ),
            instanceId: data.productionDeployment.runtimeInstanceId,
            jwt: data?.jwt,
          };
        },
        placeholderData: undefined,
      },
    }
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
{/if}
