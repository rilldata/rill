<script lang="ts">
  import { setContext } from "svelte";
  import {
    QueryClientProvider,
    type QueryClient,
  } from "@tanstack/svelte-query";
  import { RUNTIME_CONTEXT_KEY } from "../../../../runtime-client/v2/context";
  import type { RuntimeClient } from "../../../../runtime-client/v2";
  import type { V1Resource } from "../../../../runtime-client";
  import BackfillPartitionsPanel from "../BackfillPartitionsPanel.svelte";

  export let client: RuntimeClient;
  export let queryClient: QueryClient;
  export let resource: V1Resource;
  export let onTriggered: () => void;

  setContext(RUNTIME_CONTEXT_KEY, client);
</script>

<QueryClientProvider client={queryClient}>
  <BackfillPartitionsPanel {resource} {onTriggered} onDone={() => {}} />
</QueryClientProvider>
