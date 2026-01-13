<script lang="ts">
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";

  export let instanceId: string;
  export let tables: V1OlapTableInfo[] = [];
  export let rowCounts: Map<string, number | "loading" | "error"> = new Map();

  // Row count fetching logic removed temporarily
</script>

<!-- Data fetching component, renders nothing -->
