<script lang="ts">
  import {
    RUNTIME_CONTEXT_KEY,
    type RuntimeClient,
  } from "@rilldata/web-common/runtime-client/v2";
  import { setQueryClientContext } from "@tanstack/svelte-query";
  import type { QueryClient } from "@tanstack/svelte-query";
  import { setContext, untrack } from "svelte";

  /**
   * Test-only harness that supplies the contexts app code expects and runs `init`
   * during component initialization. Anything created in `init` therefore gets a
   * QueryClient for `createQuery` and an owner for its `$effect`s.
   */
  let {
    queryClient,
    runtimeClient,
    init,
  }: {
    queryClient: QueryClient;
    runtimeClient: RuntimeClient;
    init: () => void;
  } = $props();

  // The props are set once by the test and never updated, so reading them here rather
  // than in a closure is intentional.
  setQueryClientContext(untrack(() => queryClient));
  setContext(
    RUNTIME_CONTEXT_KEY,
    untrack(() => runtimeClient),
  );

  untrack(() => init)();
</script>
