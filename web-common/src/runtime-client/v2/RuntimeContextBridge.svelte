<script lang="ts">
  import { setContext } from "svelte";
  import type { RuntimeClient } from "./runtime-client";
  import { RUNTIME_CONTEXT_KEY } from "./context";

  /**
   * Injects a RuntimeClient into Svelte context so that children
   * can call useRuntimeClient(). Use this to bridge components that
   * render outside of RuntimeProvider but have access to a RuntimeClient
   * via a store or prop.
   *
   * Wrap in {#key client} to re-mount when the client instance changes.
   */
  export let client: RuntimeClient;
  setContext(RUNTIME_CONTEXT_KEY, client);
</script>

<slot />
