<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createIframeRPCHandler } from "@rilldata/web-common/lib/rpc";

  const instanceId = $page.url.searchParams.get("instance_id");
  const runtimeHost = $page.url.searchParams.get("runtime_host");
  const accessToken = $page.url.searchParams.get("access_token");

  onMount(() => {
    createIframeRPCHandler();
  });
</script>

<RuntimeProvider
  host={runtimeHost}
  {instanceId}
  jwt={accessToken}
  authContext="embed"
>
  <slot />
</RuntimeProvider>
