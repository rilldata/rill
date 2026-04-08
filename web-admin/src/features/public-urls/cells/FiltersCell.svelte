<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import type { V1Expression } from "@rilldata/web-admin/client";

  let {
    metricsViewFilters,
  }: {
    metricsViewFilters: { [key: string]: V1Expression } | undefined;
  } = $props();

  interface FilterEntry {
    name: string;
    values: string[];
    isInclude: boolean;
  }

  function extractDimensionFilters(expr: V1Expression): FilterEntry[] {
    if (!expr?.cond) return [];
    const op = expr.cond.op;
    const exprs = expr.cond.exprs ?? [];

    if (op === "OPERATION_AND" || op === "OPERATION_OR") {
      return exprs.flatMap((sub) => extractDimensionFilters(sub));
    }

    const isInclude = op !== "OPERATION_NIN" && op !== "OPERATION_NLIKE";
    const identExpr = exprs.find((e) => e.ident);
    const name = identExpr?.ident ?? "Unknown";
    const values = exprs
      .filter((e) => e.val !== undefined)
      .map((e) => String(e.val));
    return [{ name, values, isInclude }];
  }

  const filters = $derived(
    metricsViewFilters
      ? Object.values(metricsViewFilters).flatMap((expr) =>
          extractDimensionFilters(expr),
        )
      : [],
  );
</script>

{#if filters.length > 0}
  <div class="flex gap-1 flex-wrap">
    {#each filters as filter (filter.name)}
      <Chip
        type="dimension"
        readOnly
        exclude={!filter.isInclude}
        compact
        slideDuration={0}
      >
        <span slot="body" class="text-xs truncate">
          <span class="font-bold">
            {filter.isInclude ? "" : "Exclude "}{filter.name}
          </span>
          {#if filter.values.length === 1}
            {filter.values[0]}
          {:else if filter.values.length > 1}
            <span class="italic">
              {filter.values[0]} +{filter.values.length - 1}
            </span>
          {/if}
        </span>
      </Chip>
    {/each}
  </div>
{:else}
  <span class="text-fg-secondary">—</span>
{/if}
