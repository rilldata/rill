<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useReport } from "./selectors";

  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);
  $: console.log("reportQuery", $reportQuery.data);

  $: metricViewsName =
    ($reportQuery.data &&
      JSON.parse($reportQuery.data.resource.report.spec.queryArgsJson)
        ?.metrics_view_name) ??
    "";
  $: dashboard = useDashboard($runtime.instanceId, metricViewsName);
  $: dashboardTitle = $dashboard.data?.metricsView.spec.title;
</script>

{#if $reportQuery.data}
  <div class="flex flex-col gap-y-2 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-800 text-base font-medium leading-none">
          {$reportQuery.data.resource.meta.name.name}
        </h1>
        <Tag>Report</Tag>
        <div class="grow" />
        <Button type="primary">Export now</Button>
      </div>
    </div>

    <!-- three columns of metadata -->
    <div class="grid grid-cols-3 gap-x-20 grid-cols-[auto,auto,1fr]">
      <!-- column 1 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- dashboard name -->
        <span class="text-gray-800 text-sm font-medium">Dashboard</span>
        <span class="text-gray-700 text-sm font-normal">
          {dashboardTitle}
        </span>
        <!-- author -->
        <span class="text-gray-800 text-sm font-medium">Creator</span>
        <span class="text-gray-700 text-sm font-normal">
          {$reportQuery.data.resource.report.spec.annotations[
            "admin_owner_user_id"
          ] ?? "Created by a project admin"}
        </span>
        <!-- recipients -->
        <span class="text-gray-800 text-sm font-medium">Recipients</span>
        <span class="text-gray-700 text-sm font-normal">
          {$reportQuery.data.resource.report.spec.emailRecipients.length}
        </span>
        <!-- TODO -->
      </div>

      <!-- column 2 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- last run -->
        <span class="text-gray-800 text-sm font-medium">Format</span>
        <span class="text-gray-700 text-sm font-normal">
          <!-- TODO: make this pretty -->
          {$reportQuery.data.resource.report.spec.exportFormat}
        </span>
        <!-- next run -->
        <span class="text-gray-800 text-sm font-medium">Limit</span>
        <span class="text-gray-700 text-sm font-normal">
          {$reportQuery.data.resource.report.spec.exportLimit}
        </span>
      </div>

      <!-- column 3 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- last run -->
        <span class="text-gray-800 text-sm font-medium">Frequency</span>
        <span class="text-gray-700 text-sm font-normal">
          {$reportQuery.data.resource.report.spec.refreshSchedule.cron}
        </span>
        <!-- next run -->
        <span class="text-gray-800 text-sm font-medium">Next run</span>
        <span class="text-gray-700 text-sm font-normal">
          {$reportQuery.data.resource.report.state.nextRunOn}
        </span>
      </div>
    </div>
  </div>
{/if}
