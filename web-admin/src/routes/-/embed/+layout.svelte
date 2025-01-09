<script lang="ts">
  import { page } from "$app/stores";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";

  const instanhceId = $page.url.searchParams.get("instance_id");
  const runtimeHost = $page.url.searchParams
    .get("runtime_host")
    .replace("localhost:9091", "localhost:8081");
  const accessToken = $page.url.searchParams.get("access_token");
</script>

<RuntimeProvider
  host="localhost:8081"
  instanceId="03000859377c4b33a8fe18062dfa1dca"
  jwt={accessToken}
  authContext="embed"
>
  <slot />
</RuntimeProvider>
