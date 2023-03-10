<script lang="ts">
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";

  // TODO: call the admin server
  $: runtimeHost = undefined;
  $: runtimeInstanceId = undefined;
  $: jwt = undefined;
</script>

<RuntimeProvider host={runtimeHost} instanceId={runtimeInstanceId} {jwt}>
  <slot />
</RuntimeProvider>
