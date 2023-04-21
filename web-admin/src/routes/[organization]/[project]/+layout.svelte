<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createAdminServiceGetProject } from "../../../client";

  $: proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project,
    {
      query: {
        // Proactively refetch the JWT because it's only valid for 1 hour
        refetchInterval: 1000 * 60 * 30, // 30 minutes
      },
    }
  );

  // Hack: in development, the runtime host is actually on port 8081
  $: runtimeHost = $proj.data?.productionDeployment?.runtimeHost.replace(
    "localhost:9091",
    "localhost:8081"
  );
  $: runtimeInstanceId = $proj.data?.productionDeployment?.runtimeInstanceId;
  $: jwt = $proj.data?.jwt;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
