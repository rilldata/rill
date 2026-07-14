<script lang="ts">
  import type { QueryObserverResult } from "@tanstack/svelte-query";
  import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
  import type { ConnectError } from "@connectrpc/connect";
  import type { Readable } from "svelte/store";

  export let validSpecQuery: Readable<
    QueryObserverResult<ExploreValidSpecResponse, ConnectError>
  >;
</script>

{!!$validSpecQuery.data}
