<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useReport } from "../selectors";
  import MetadataLabel from "./MetadataLabel.svelte";
  import MetadataValue from "./MetadataValue.svelte";

  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);

  $: metricViewsName =
    ($reportQuery.data &&
      JSON.parse($reportQuery.data.resource.report.spec.queryArgsJson)
        ?.metrics_view_name) ??
    "";
  $: dashboard = useDashboard($runtime.instanceId, metricViewsName);
  $: dashboardTitle = $dashboard.data?.metricsView.spec.title;
</script>

{#if $reportQuery.data}
  <div class="flex flex-col gap-y-8 w-full max-w-full 2xl:max-w-[1200px]">
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

    <!-- Three columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-3">
      <!-- column 1 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- dashboard name -->
        <MetadataLabel>Dashboard</MetadataLabel>
        <MetadataValue>
          {dashboardTitle}
        </MetadataValue>
        <!-- author -->
        <MetadataLabel>Creator</MetadataLabel>
        <MetadataValue>
          {$reportQuery.data.resource.report.spec.annotations[
            "admin_owner_user_id"
          ] ?? "Created by a project admin"}
        </MetadataValue>
        <!-- recipients -->
        <MetadataLabel>Recipients</MetadataLabel>
        <MetadataValue>
          {$reportQuery.data.resource.report.spec.emailRecipients.length}
        </MetadataValue>
      </div>

      <!-- column 2 -->
      <div class="flex flex-col gap-y-3">
        <!-- last run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Format</MetadataLabel>
          <MetadataValue>
            <!-- TODO: make this pretty -->
            {$reportQuery.data.resource.report.spec.exportFormat}
          </MetadataValue>
        </div>
        <!-- next run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Limit</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data.resource.report.spec.exportLimit}
          </MetadataValue>
        </div>
      </div>

      <!-- column 3 -->
      <div class="flex flex-col gap-y-3">
        <!-- last run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Frequency</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data.resource.report.spec.refreshSchedule.cron}
          </MetadataValue>
        </div>
        <!-- next run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Next run</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data.resource.report.state.nextRunOn}
          </MetadataValue>
        </div>
      </div>
    </div>
  </div>
{/if}
