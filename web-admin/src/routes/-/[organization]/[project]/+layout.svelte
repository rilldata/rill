<script lang="ts">
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";

  // TODO: call the admin server
  $: runtimeHost = "http://localhost:9009";
  $: runtimeInstanceId = "default";
  $: jwt = undefined;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
