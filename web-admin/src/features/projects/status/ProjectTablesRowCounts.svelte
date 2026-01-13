<script lang="ts">
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fetchRowCount } from "./selectors";

  export let instanceId: string;
  export let tables: V1OlapTableInfo[] = [];
  export let rowCounts: Map<string, number | "loading" | "error"> = new Map();

  // Track which tables we've already fetched
  let fetchedTables = new Set<string>();

  // Fetch row counts when JWT is ready and tables change
  $: if (instanceId && tables.length > 0 && $runtime?.jwt?.token && $runtime.jwt.token !== "") {
    for (const table of tables) {
      const tableName = table.name;
      if (!tableName || fetchedTables.has(tableName)) continue;

      fetchedTables.add(tableName);
      rowCounts.set(tableName, "loading");
      rowCounts = rowCounts;

      // Fetch row count
      fetchRowCount(instanceId, tableName).then((count) => {
        rowCounts.set(tableName, count);
        rowCounts = rowCounts;
      });
    }
  }
</script>

<!-- Data fetching component, renders nothing -->
