<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { formatRunDate } from "../tableUtils";
  import cronstrue from "cronstrue";
  import ReportOwnerBullet from "./ReportOwnerBullet.svelte";

  export let data: V1Resource[];
  export let organization: string;
  export let project: string;

  let searchText = "";

  $: filteredData = data.filter((row) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    const name = (row.report?.spec?.displayName || "").toLowerCase();
    const dashboard = getDashboardName(row).toLowerCase();
    return name.includes(q) || dashboard.includes(q);
  });

  function getDashboardName(row: V1Resource): string {
    try {
      return getDashboardNameFromReport(row.report?.spec) ?? "";
    } catch {
      return "";
    }
  }

  function getLastRun(row: V1Resource): string | undefined {
    return row.report?.state?.executionHistory?.[0]?.reportTime;
  }

  function getLastRunError(row: V1Resource): string | undefined {
    return row.report?.state?.executionHistory?.[0]?.errorMessage;
  }

  function getFrequency(row: V1Resource): string {
    const cron = row.report?.spec?.refreshSchedule?.cron;
    if (!cron) return "—";
    try {
      return cronstrue.toString(cron);
    } catch {
      return cron;
    }
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
          icon={ReportIcon}
          message="You don't have any reports yet"
        >
          <span slot="action">
            Schedule <a
              href="https://docs.rilldata.com/guide/reports/exports"
              target="_blank"
              rel="noopener noreferrer"
            >
              reports</a
            > from any dashboard
          </span>
        </ResourceListEmptyState>
      </div>
    </div>
  {:else if filteredData.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16 text-fg-secondary text-sm font-semibold">
        No reports match your search
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
            <th class="table-header">Last run</th>
            <th class="table-header">Schedule</th>
            <th class="table-header">Created by</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredData as row (row.meta?.name?.name)}
            {@const id = row.meta?.name?.name ?? ""}
            {@const title = row.report?.spec?.displayName || id}
            {@const lastRun = getLastRun(row)}
            {@const errorMessage = getLastRunError(row)}
            {@const dashboard = getDashboardName(row)}
            {@const timeZone =
              row.report?.spec?.refreshSchedule?.timeZone ?? "UTC"}
            {@const ownerId =
              row.report?.spec?.annotations?.["admin_owner_user_id"] ?? ""}
            <tr class="table-row">
              <td class="table-cell font-medium">
                <a
                  href={`reports/${id}`}
                  class="flex items-center gap-x-2 hover:text-accent-primary-action"
                >
                  <ReportIcon size="14px" />
                  <span class="truncate">{title}</span>
                </a>
              </td>
              <td class="table-cell">
                <span class="truncate block">{dashboard || "—"}</span>
              </td>
              <td class="table-cell">
                {#if !lastRun}
                  <span class="text-fg-secondary">—</span>
                {:else if errorMessage}
                  <CancelCircleInverse className="text-red-500" />
                {:else}
                  <CheckCircleOutline className="text-primary-500" />
                {/if}
              </td>
              <td class="table-cell">
                {#if lastRun}
                  <span class="whitespace-nowrap"
                    >{formatRunDate(lastRun, timeZone)}</span
                  >
                {:else}
                  <span class="text-fg-secondary">Never</span>
                {/if}
              </td>
              <td class="table-cell">
                <span class="truncate block">{getFrequency(row)}</span>
              </td>
              <td class="table-cell">
                <ReportOwnerBullet {organization} {project} {ownerId} />
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
