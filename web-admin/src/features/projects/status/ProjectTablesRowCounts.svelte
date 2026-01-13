<script lang="ts">
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";

  /**
   * Props are used by parent component via two-way binding.
   * Suppress unused-export-let warnings as these are intentionally for external reference.
   */
  // @ts-ignore unused-export-let
  export let instanceId: string;
  // @ts-ignore unused-export-let
  export let tables: V1OlapTableInfo[] = [];
  // @ts-ignore unused-export-let
  export let rowCounts: Map<string, number | "loading" | "error"> = new Map();

  // Row count fetching logic removed temporarily
  // These variables are captured in closure for later when fetching is re-enabled
  void instanceId;
  void tables;
  void rowCounts;
</script>

<!-- Data fetching component, renders nothing -->
