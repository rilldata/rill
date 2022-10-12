<script lang="ts">
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import RillTheme from "@rilldata/web-local/lib/components/layouts/RillTheme.svelte";
  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();
</script>

<RillTheme>
  <slot />
</RillTheme>
