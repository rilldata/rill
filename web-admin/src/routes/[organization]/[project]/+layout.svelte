<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";

  const org = $page.params.organization;
  const project = $page.params.project;

  // TODO: call the admin server
  $: runtimeHost = "http://localhost:9009";
  $: runtimeInstanceId = "default";
  $: jwt = undefined;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
