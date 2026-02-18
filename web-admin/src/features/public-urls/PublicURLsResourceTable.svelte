<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { Search } from "@rilldata/web-common/components/search";
  import { ExternalLinkIcon } from "lucide-svelte";
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
  import type { V1Expression } from "@rilldata/web-admin/client";
  import PublicURLsActionsRow from "./PublicURLsActionsRow.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";

  interface PublicURLRow extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  export let data: PublicURLRow[];
  export let onDelete: (deletedTokenId: string) => void;

  let searchText = "";

  $: filteredData = data.filter((row) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    const label = (row.displayName || row.dashboardTitle || "").toLowerCase();
    const dashboard = (row.dashboardTitle || "").toLowerCase();
    const creator = String(row.attributes?.name || "").toLowerCase();
    return label.includes(q) || dashboard.includes(q) || creator.includes(q);
  });

  function formatDate(value: string | undefined) {
    if (!value) return "—";
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

  function getFilters(
    filters: { [key: string]: V1Expression } | undefined,
  ): FilterEntry[] {
    if (!filters) return [];
    return Object.values(filters).flatMap((expr) =>
      extractDimensionFilters(expr),
    );
  }
</script>

<div class="flex flex-col gap-y-3 w-full">
  <div class="w-64">
    <Search
      placeholder="Search"
      autofocus={false}
      bind:value={searchText}
      rounded="lg"
    />
  </div>

  {#if filteredData.length === 0 && data.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16">
        <ResourceListEmptyState
          icon={ExternalLinkIcon}
          message="You don't have any public URLs yet"
        >
          <span slot="action">
            To create a public URL, click the Share button in a dashboard.
          </span>
        </ResourceListEmptyState>
      </div>
    </div>
  {:else if filteredData.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16 text-fg-secondary text-sm font-semibold">
        No public URLs match your search
      </div>
    </div>
  {:else}
    <div class="border rounded-lg overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="bg-surface-background border-b">
            <th class="table-header">Label</th>
            <th class="table-header">Dashboard</th>
            <th class="table-header">Filters</th>
            <th class="table-header">Expires on</th>
            <th class="table-header">Created by</th>
            <th class="table-header">Last accessed</th>
            <th class="table-header w-12"></th>
          </tr>
        </thead>
        <tbody>
          {#each filteredData as row (row.id)}
            {@const filters = getFilters(row.metricsViewFilters)}
            <tr class="table-row">
              <td class="table-cell font-medium">
                <a
                  href={row.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-x-2 hover:text-accent-primary-action"
                >
                  <ExternalLinkIcon size={14} class="shrink-0" />
                  <span class="truncate">
                    {row.displayName || row.dashboardTitle || "Untitled"}
                  </span>
                </a>
              </td>
              <td class="table-cell">
                <span class="truncate block">
                  {row.dashboardTitle || row.resourceName || "—"}
                </span>
              </td>
              <td class="table-cell">
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
                          <span class="font-bold"
                            >{filter.isInclude
                              ? ""
                              : "Exclude "}{filter.name}</span
                          >
                          {#if filter.values.length === 1}
                            {filter.values[0]}
                          {:else if filter.values.length > 1}
                            <span class="italic"
                              >{filter.values[0]} +{filter.values.length -
                                1}</span
                            >
                          {/if}
                        </span>
                      </Chip>
                    {/each}
                  </div>
                {:else}
                  <span class="text-fg-secondary">—</span>
                {/if}
              </td>
              <td class="table-cell">
                {formatDate(row.expiresOn)}
              </td>
              <td class="table-cell">
                {row.attributes?.name || "—"}
              </td>
              <td class="table-cell">
                {formatDate(row.usedOn)}
              </td>
              <td class="table-cell w-12">
                <PublicURLsActionsRow
                  id={row.id ?? ""}
                  url={row.url ?? ""}
                  {onDelete}
                />
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style lang="postcss">
  .table-header {
    @apply px-4 py-2.5 text-left text-xs font-medium text-fg-secondary whitespace-nowrap;
  }

  .table-row {
    @apply border-b bg-surface-background;
  }

  .table-row:last-child {
    @apply border-b-0;
  }

  .table-row:hover {
    @apply bg-surface-hover;
  }

  .table-cell {
    @apply px-4 py-2.5 text-sm text-fg-primary;
  }
</style>
