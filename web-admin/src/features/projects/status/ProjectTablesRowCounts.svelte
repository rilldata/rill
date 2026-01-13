<script lang="ts">
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { httpClient } from "@rilldata/web-common/runtime-client/http-client";

  export let instanceId: string;
  export let tables: V1OlapTableInfo[] = [];
  export let rowCounts: Map<string, number | "loading" | "error"> = new Map();

  // Track which tables we've already queried
  let queriedTables = new Set<string>();

  // Reset queriedTables when JWT changes (in case token was refreshed)
  $: if (!$runtime?.jwt?.token) {
    queriedTables.clear();
  }

  // Fetch row counts for all tables (only when JWT is ready)
  $: if (instanceId && tables && tables.length > 0 && $runtime?.jwt?.token && $runtime.jwt.token !== "") {
    for (const table of tables) {
      const tableName = table.name;
      if (!tableName || queriedTables.has(tableName)) continue;

      queriedTables.add(tableName);
      rowCounts.set(tableName, "loading");
      rowCounts = rowCounts;

      // Fetch row count directly with retry logic
      const fetchCount = async () => {
        let retryCount = 0;
        const maxRetries = 3;

        const attemptFetch = async (attempt: number) => {
          try {
            const waitTime = attempt === 0 ? 300 : 1500;
            console.log(`[RowCount] Attempt ${attempt + 1}/${maxRetries + 1} for ${tableName}, waiting ${waitTime}ms`);
            await new Promise(resolve => setTimeout(resolve, waitTime));

            console.log(`[RowCount] Calling httpClient for ${tableName}...`);
            const response_body = await httpClient({
              url: `/v1/instances/${instanceId}/query`,
              method: "POST",
              headers: { "Content-Type": "application/json" },
              data: {
                sql: `SELECT COUNT(*) as count FROM "${tableName}"`,
              },
            });

            console.log(`[RowCount] Success for ${tableName}:`, response_body);

            // Extract count from response
            if (response_body?.data && Array.isArray(response_body.data) && response_body.data.length > 0) {
              const firstRow = response_body.data[0] as any;
              const count = parseInt(String(firstRow?.count ?? 0), 10);
              console.log(`[RowCount] Extracted count for ${tableName}:`, count);
              rowCounts.set(tableName, isNaN(count) ? "error" : count);
            } else {
              console.warn(`[RowCount] Invalid response structure for ${tableName}:`, response_body);
              rowCounts.set(tableName, "error");
            }
            rowCounts = rowCounts;
          } catch (error: any) {
            const status = error?.response?.status || error?.status;
            const message = error?.message || String(error);
            console.error(`[RowCount] Attempt ${attempt + 1} failed for ${tableName} (status: ${status}):`, message);

            if ((status === 401 || message.includes("401")) && attempt < maxRetries) {
              console.log(`[RowCount] Got 401, retrying ${tableName}...`);
              await attemptFetch(attempt + 1);
            } else {
              console.error(`[RowCount] Final error for ${tableName} after ${attempt + 1} attempts`);
              rowCounts.set(tableName, "error");
              rowCounts = rowCounts;
            }
          }
        };

        await attemptFetch(0);
      };

      void fetchCount();
    }
  }
</script>

<!-- This is a data-fetching component, it doesn't render anything -->
