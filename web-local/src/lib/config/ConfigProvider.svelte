<script lang="ts">
  import { RootConfig } from "$web-local/common/config/RootConfig";
  import { ServerConfig } from "$web-local/common/config/ServerConfig";
  import { setContext } from "svelte";

  export let config: RootConfig;
  export let exploreUrl: string;

  setContext(
    "config",
    config
      ? config
      : new RootConfig({
          ...(exploreUrl
            ? {
                server: new ServerConfig({
                  exploreUrl,
                }),
              }
            : {}),
        })
  );
</script>

<slot />
