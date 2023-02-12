<script lang="ts">
  import { RillTheme } from "@rilldata/web-common/layout";
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();
</script>

<RillTheme>
  <slot />
</RillTheme>
