<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import { mapQueryToDashboard } from "@rilldata/web-admin/features/dashboards/query-mappers/mapQueryToDashboard";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertId = $page.params.alert;
  $: executionTime = $page.url.searchParams.get("execution_time");

  $: alert = useAlert($runtime.instanceId, alertId);

  let dashboardStateForReport: ReturnType<typeof mapQueryToDashboard>;
  $: dashboardStateForReport = mapQueryToDashboard(
    $alert.data?.resource?.alert?.spec?.queryName ?? "",
    $alert.data?.resource?.alert?.spec?.queryArgsJson ?? "",
    executionTime,
  );

  $: if ($dashboardStateForReport.data) {
    goto(
      `/${organization}/${project}/${$dashboardStateForReport.data.metricsView}?state=${encodeURIComponent($dashboardStateForReport.data.state)}`,
    );
  }
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
        on:click={() => goto(`/${organization}/${project}/-/alerts/${$alert}`)}
        variant="primary-outline"
      >
        Go to Alerts page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
