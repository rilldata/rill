<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import { timeAgo } from "../../dashboards/listing/utils";
  import { getAlertDashboardName } from "../selectors";
  import AlertOwnerBullet from "./AlertOwnerBullet.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";

  $: filteredData = data.filter((row) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    const name = (
      row.alert?.spec?.displayName ||
      row.meta?.name?.name ||
      ""
    ).toLowerCase();
    const dashboard = getAlertDashboardName(row.alert?.spec).toLowerCase();
    return name.includes(q) || dashboard.includes(q);
  });

  function getLastTrigger(row: V1Resource): string | undefined {
    return (
      row.alert?.state?.executionHistory?.[0]?.finishedOn ??
      row.alert?.state?.executionHistory?.[0]?.startedOn
    );
  }

  function getLastTriggerError(row: V1Resource): string | undefined {
    return row.alert?.state?.executionHistory?.[0]?.result?.errorMessage;
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
          icon={AlertIcon}
          message="You don't have any alerts yet"
        >
          <span slot="action">
            Create <a
              href="https://docs.rilldata.com/guide/alerts"
              target="_blank"
              rel="noopener noreferrer"
            >
              alerts
            </a>
            from any dashboard or{" "}
            <a
              href="https://docs.rilldata.com/reference/project-files/alerts"
              target="_blank"
              rel="noopener noreferrer"
            >
              via code</a
            >.
          </span>
        </ResourceListEmptyState>
      </div>
    </div>
  {:else if filteredData.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16 text-fg-secondary text-sm font-semibold">
        No alerts match your search
      </div>
    </div>
  {:else}
    <div class="border rounded-lg overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="bg-surface-background border-b">
            <th class="table-header">Label</th>
            <th class="table-header">Dashboard</th>
            <th class="table-header">Status</th>
            <th class="table-header">Last checked</th>
            <th class="table-header">Created by</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredData as row (row.meta?.name?.name)}
            {@const id = row.meta?.name?.name ?? ""}
            {@const title = row.alert?.spec?.displayName || id}
            {@const lastTrigger = getLastTrigger(row)}
            {@const errorMessage = getLastTriggerError(row)}
            {@const dashboard = getAlertDashboardName(row.alert?.spec)}
            {@const ownerId =
              row.alert?.spec?.annotations?.["admin_owner_user_id"] ?? ""}
            <tr class="table-row">
              <td class="table-cell font-medium">
                <a
                  href={`alerts/${id}`}
                  class="flex items-center gap-x-2 hover:text-accent-primary-action"
                >
                  <AlertIcon size="14px" />
                  <span class="truncate">{title}</span>
                </a>
              </td>
              <td class="table-cell">
                <span class="truncate block">{dashboard || "—"}</span>
              </td>
              <td class="table-cell">
                {#if !lastTrigger}
                  <span class="text-fg-secondary">—</span>
                {:else if errorMessage}
                  <CancelCircleInverse className="text-red-500" />
                {:else}
                  <CheckCircleOutline className="text-primary-500" />
                {/if}
              </td>
              <td class="table-cell">
                {#if lastTrigger}
                  <span>{timeAgo(new Date(lastTrigger))}</span>
                {:else}
                  <span class="text-fg-secondary">Never</span>
                {/if}
              </td>
              <td class="table-cell">
                <AlertOwnerBullet {organization} {project} {ownerId} />
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
