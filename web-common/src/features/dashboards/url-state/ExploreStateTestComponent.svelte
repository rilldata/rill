<script lang="ts">
  import type { QueryObserverResult } from "@rilldata/svelte-query";
  import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
  import type { RpcStatus } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";

  export let validSpecQuery: Readable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  >;
</script>

{!!$validSpecQuery.data}
