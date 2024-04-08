<script lang="ts">
  import { RillTheme } from "@rilldata/web-common/layout";
  import { addViewportListener } from "@rilldata/web-common/lib/viewport-utils";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let data;

  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();

  addViewportListener();

  runtime.set({
    host: data.host,
    instanceId: data.instanceId,
  });
</script>

<RillTheme>
  <slot />
</RillTheme>
