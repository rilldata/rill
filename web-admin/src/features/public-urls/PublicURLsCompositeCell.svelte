<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import type { V1Expression } from "@rilldata/web-admin/client";
  import { ExternalLinkIcon } from "lucide-svelte";
  import PublicURLsActionsRow from "./PublicURLsActionsRow.svelte";

  export let url: string;
  export let displayName: string;
  export let dashboardTitle: string;
  export let createdBy: string;
  export let expiresOn: string | undefined;
  export let id: string;
  export let onDelete: (deletedTokenId: string) => void;
  export let metricsViewFilters:
    | { [key: string]: V1Expression }
    | undefined = undefined;

  function formatDate(value: string) {
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  interface FilterEntry {
    name: string;
    values: string[];
    isInclude: boolean;
  }

  function extractDimensionFilters(expr: V1Expression): FilterEntry[] {
    if (!expr?.cond) return [];
    const op = expr.cond.op;
    const exprs = expr.cond.exprs ?? [];

    // Compound expression (AND/OR) — recurse into sub-expressions
    if (op === "OPERATION_AND" || op === "OPERATION_OR") {
      return exprs.flatMap((sub) => extractDimensionFilters(sub));
    }

    // Leaf filter (IN/NIN/LIKE/NLIKE etc.) — first expr is the dimension ident, rest are values
    const isInclude = op !== "OPERATION_NIN" && op !== "OPERATION_NLIKE";
    const identExpr = exprs.find((e) => e.ident);
    const name = identExpr?.ident ?? "Unknown";
    const values = exprs
      .filter((e) => e.val !== undefined)
      .map((e) => String(e.val));
    return [{ name, values, isInclude }];
  }

  $: filterEntries = metricsViewFilters
    ? Object.values(metricsViewFilters).flatMap((expr) =>
        extractDimensionFilters(expr),
      )
    : [];
</script>

<div class="flex items-center justify-between px-4 py-2.5 w-full h-full">
  <a
    href={url}
    target="_blank"
    rel="noopener noreferrer"
    class="flex flex-col gap-y-1 group flex-1 min-w-0"
  >
    <div class="flex gap-x-2 items-center min-h-[20px]">
      <ExternalLinkIcon size={14} />
      <span
        class="text-fg-primary text-sm font-semibold group-hover:text-accent-primary-action truncate"
      >
        {displayName || dashboardTitle || "Untitled"}
      </span>
    </div>
    <div
      class="flex gap-x-1 text-fg-secondary text-xs font-normal min-h-[16px] overflow-hidden items-center"
    >
      {#if filterEntries.length > 0}
        {#each filterEntries as filter (filter.name)}
          <Chip
            type="dimension"
            readOnly
            exclude={!filter.isInclude}
            compact
            slideDuration={0}
          >
            <span slot="body" class="text-xs truncate">
              <span class="font-bold"
                >{filter.isInclude ? "" : "Exclude "}{filter.name}</span
              >
              {#if filter.values.length === 1}
                {filter.values[0]}
              {:else if filter.values.length > 1}
                <span class="italic"
                  >{filter.values[0]} +{filter.values.length - 1}
                  other{filter.values.length - 1 !== 1 ? "s" : ""}</span
                >
              {/if}
            </span>
          </Chip>
        {/each}
      {/if}
      {#if createdBy}
        <span class="shrink-0">• Created by </span>
        <span class="shrink-0">{createdBy}</span>
      {/if}
      {#if expiresOn}
        <span class="shrink-0">•</span>
        <span class="shrink-0">Expires {formatDate(expiresOn)}</span>
      {/if}
    </div>
  </a>
  <div class="shrink-0">
    <PublicURLsActionsRow {id} {url} {onDelete} />
  </div>
</div>
