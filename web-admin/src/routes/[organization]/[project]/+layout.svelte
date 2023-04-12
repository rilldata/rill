<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createAdminServiceGetProject } from "../../../client";

  $: proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );

  // Hack: in development, the runtime host is actually on port 8081
  $: runtimeHost = $proj.data?.productionDeployment?.runtimeHost.replace(
    "localhost:9091",
    "localhost:8081"
  );
  $: runtimeInstanceId = $proj.data?.productionDeployment?.runtimeInstanceId;
  $: jwt = $proj.data.jwt;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
