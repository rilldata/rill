<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getDashboardStateForReport } from "@rilldata/web-admin/features/scheduled-reports/get-dashboard-state-for-report";
  import { useReport } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportId = $page.params.report;
  $: executionTime = $page.url.searchParams.get("execution_time");

  $: report = useReport($runtime.instanceId, reportId);

  let dashboardStateForReport: ReturnType<typeof getDashboardStateForReport>;
  $: dashboardStateForReport = getDashboardStateForReport(
    $report.data?.resource,
    executionTime,
  );

  $: if ($dashboardStateForReport.data) {
    goto(
      `/${organization}/${project}/${$dashboardStateForReport.data.metricsView}?state=${$dashboardStateForReport.data.state}`,
    );
  }

  // TODO: error handling
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $dashboardStateForReport.isFetching}
      <div class="h-36 mt-10">
        <Spinner
          bg="linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)"
          status={EntityStatus.Running}
          size="7rem"
          duration={725}
        />
      </div>
    {:else if $dashboardStateForReport.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to open report</h2>
        <CtaMessage>
          {$dashboardStateForReport.error}
        </CtaMessage>
      </div>
      <CtaButton
        on:click={() =>
          goto(`/${organization}/${project}/-/reports/${reportId}`)}
        variant="primary-outline"
      >
        Go to report page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
