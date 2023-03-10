<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { useAdminServiceGetProject } from "../../../../client";

  const proj = useAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );

  $: runtimeHost = $proj.data.productionDeployment.runtimeHost;
  $: runtimeInstanceId = $proj.data.productionDeployment.runtimeInstanceId;
  $: jwt = $proj.data.jwt;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
