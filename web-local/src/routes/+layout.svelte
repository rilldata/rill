<script lang="ts">
  import { initializeNodeStoreContexts } from "@rilldata/web-local/lib/layouts";
  /** This function will initialize the existing node stores and will connect them
   * to the Node server.
   */
  initializeNodeStoreContexts();
</script>

<slot />
