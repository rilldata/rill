<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-to-explore";
  import {
    getExploreName,
    getExplorePageUrlSearchParams,
  } from "@rilldata/web-common/features/explore-mappers/utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertId = $page.params.alert;
  $: executionTime = $page.url.searchParams.get("execution_time");

  $: alert = useAlert(instanceId, alertId);
  $: exploreName = getExploreName(
    $alert.data?.resource?.alert?.spec?.annotations?.web_open_path,
  );

  let dashboardStateForAlert: ReturnType<typeof mapQueryToDashboard>;
  $: queryName =
    ($alert.data?.resource?.alert?.spec?.resolverProperties?.query_name as
      | string
      | undefined) ??
    $alert.data?.resource?.alert?.spec?.queryName ??
    "";
  $: queryArgsJson =
    ($alert.data?.resource?.alert?.spec?.resolverProperties?.query_args_json as
      | string
      | undefined) ??
    $alert.data?.resource?.alert?.spec?.queryArgsJson ??
    "";
  $: dashboardStateForAlert = mapQueryToDashboard({
    exploreName,
    queryName,
    queryArgsJson,
    executionTime,
    annotations: $alert.data?.resource?.alert?.spec?.annotations ?? {},
  });

  $: if ($alert.data?.resource?.alert?.spec && (!queryName || !queryArgsJson)) {
    goto(`/${organization}/${project}/-/alerts/${alertId}`);
  }

  $: if ($dashboardStateForAlert?.data) {
    void gotoExplorePage();
  }

  async function gotoExplorePage() {
    const url = new URL(
      `/${organization}/${project}/explore/${encodeURIComponent(exploreName)}`,
      window.location.origin,
    );
    url.search = (
      await getExplorePageUrlSearchParams(
        $dashboardStateForAlert.data.exploreName,
        $dashboardStateForAlert.data.exploreState,
      )
    ).toString();
    return goto(url.toString());
  }

  // When alert is loading we will have a "missing require field" error in dashboardStateForAlert so check loading for both queries.
  $: loading = $alert.isPending || $dashboardStateForAlert?.isLoading;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if loading}
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
