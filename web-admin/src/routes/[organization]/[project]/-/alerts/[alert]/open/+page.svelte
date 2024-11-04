<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import { mapQueryToDashboard } from "@rilldata/web-admin/features/dashboards/query-mappers/mapQueryToDashboard";
  import {
    getExploreName,
    getExplorePageUrl,
  } from "@rilldata/web-admin/features/dashboards/query-mappers/utils.js";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertId = $page.params.alert;
  $: executionTime = $page.url.searchParams.get("execution_time");

  $: alert = useAlert($runtime.instanceId, alertId);
  $: exploreName = getExploreName(
    $alert.data?.resource?.alert?.spec?.annotations?.web_open_path,
  );

  let dashboardStateForAlert: ReturnType<typeof mapQueryToDashboard>;
  $: queryName =
    $alert.data?.resource?.alert?.spec?.resolverProperties?.query_name ??
    $alert.data?.resource?.alert?.spec?.queryName ??
    "";
  $: queryArgsJson =
    $alert.data?.resource?.alert?.spec?.resolverProperties?.query_args_json ??
    $alert.data?.resource?.alert?.spec?.queryArgsJson ??
    "";
  $: dashboardStateForAlert = mapQueryToDashboard(
    exploreName,
    queryName,
    queryArgsJson,
    executionTime,
    $alert.data?.resource?.alert?.spec?.annotations ?? {},
  );

  $: if ($alert.data?.resource?.alert?.spec && (!queryName || !queryArgsJson)) {
    goto(`/${organization}/${project}/-/alerts/${alertId}`);
  }

  async function gotExplorePage() {
    return goto(
      await getExplorePageUrl(
        $page.url,
        organization,
        project,
        $dashboardStateForAlert.data.exploreName,
        $dashboardStateForAlert.data.dashboard,
      ),
    );
  }

  $: if ($dashboardStateForAlert.data) {
    void gotExplorePage();
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $dashboardStateForAlert.isFetching}
      <div class="h-36 mt-10">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else if $dashboardStateForAlert.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to open Alert</h2>
        <CtaMessage>
          {$dashboardStateForAlert.error}
        </CtaMessage>
      </div>
      <CtaButton
        variant="secondary"
        href={`/${organization}/${project}/-/alerts/${$alert}`}
      >
        Go to Alerts page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
