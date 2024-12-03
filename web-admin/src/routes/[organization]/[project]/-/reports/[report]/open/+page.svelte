<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { mapQueryToDashboard } from "@rilldata/web-admin/features/dashboards/query-mappers/mapQueryToDashboard";
  import {
    getExploreName,
    getExplorePageUrl,
  } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
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
  $: exploreName = getExploreName(
    $report?.data?.resource?.report?.spec?.annotations?.web_open_path,
  );

  let dashboardStateForReport: ReturnType<typeof mapQueryToDashboard>;
  $: dashboardStateForReport = mapQueryToDashboard(
    exploreName,
    $report?.data?.resource?.report?.spec?.queryName,
    $report?.data?.resource?.report?.spec?.queryArgsJson,
    executionTime,
    $report?.data?.resource?.report?.spec?.annotations ?? {},
  );

  async function gotExplorePage() {
    return goto(
      await getExplorePageUrl(
        $page.url,
        organization,
        project,
        $dashboardStateForReport.data.exploreName,
        $dashboardStateForReport.data.exploreState,
      ),
    );
  }

  $: if ($dashboardStateForReport?.data) {
    void gotExplorePage();
  }

  // TODO: error handling
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $dashboardStateForReport.isFetching}
      <div class="h-36 mt-10">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else if $dashboardStateForReport.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to open report</h2>
        <CtaMessage>
          {$dashboardStateForReport.error}
        </CtaMessage>
      </div>
      <CtaButton
        href={`/${organization}/${project}/-/reports/${reportId}`}
        variant="secondary"
      >
        Go to report page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
