<script lang="ts">
  import { goto } from "$app/navigation";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-to-explore";
  import { getExplorePageUrlSearchParams } from "@rilldata/web-common/features/explore-mappers/utils";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({
    alert: alertResource,
    organization,
    project,
    alertId,
    executionTime,
    token,
    exploreName,
  } = data);

  $: alertSpec = alertResource.alert.spec;

  let dashboardStateForAlert: ReturnType<typeof mapQueryToDashboard>;
  $: queryName =
    (alertSpec?.resolverProperties?.query_name as string | undefined) ??
    alertSpec?.queryName ??
    "";
  $: queryArgsJson =
    (alertSpec?.resolverProperties?.query_args_json as string | undefined) ??
    alertSpec?.queryArgsJson ??
    "";
  $: dashboardStateForAlert = mapQueryToDashboard(
    {
      exploreName,
      queryName,
      queryArgsJson,
      executionTime,
    },
    {
      exploreProtoState: alertSpec?.annotations?.web_open_state,
      // When opening an alert from a link with token we should remove the filters from request.
      // The filters are already baked into the token, each query will have them added in the backend.
      // So adding them again will essentially apply filters twice. It will lead to incorrect results for threshold filters.
      ignoreFilters: !!token,
    },
  );

  $: if (alertSpec && (!queryName || !queryArgsJson)) {
    goto(`/${organization}/${project}/-/alerts/${alertId}`);
  }

  $: if ($dashboardStateForAlert?.data) {
    void gotoExplorePage();
  }

  async function gotoExplorePage() {
    const exploreStateParams = await getExplorePageUrlSearchParams(
      $dashboardStateForAlert.data.exploreName,
      $dashboardStateForAlert.data.exploreState,
    );

    const url = new URL(window.location.origin);
    if (token) {
      url.pathname = `/${organization}/${project}/-/share/${token}/explore/${exploreName}`;
    } else {
      url.pathname = `/${organization}/${project}/explore/${exploreName}`;
    }

    url.search = exploreStateParams.toString();
    return goto(url.toString());
  }

  // When alert is loading we will have a "missing require field" error in dashboardStateForAlert so check loading for both queries.
  $: loading = $dashboardStateForAlert?.isLoading;
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
        href={`/${organization}/${project}/-/alerts/${alertId}`}
      >
        Go to Alerts page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
