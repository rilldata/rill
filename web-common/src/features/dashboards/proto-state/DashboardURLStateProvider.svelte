<script lang="ts">
  import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";
  import { type Unsubscriber } from "svelte/store";
  import ErrorPage from "../../../components/ErrorPage.svelte";
  import type { HTTPError } from "../../../runtime-client/fetchWrapper";
  import { getStateManagers } from "../state-managers/state-managers";

  export let metricsViewName: string;

  const ctx = getStateManagers();
  let unsubscribe: Unsubscriber;
  const {
    runtime,
    metricsViewName: ctxName,
    dashboardStore,
    timeRangeSummaryStore,
  } = ctx;
  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    $runtime.instanceId,
    $ctxName,
  );
  $: ({ error: schemaError } = $metricsViewSchema);

  $: ({ data, error } = $timeRangeSummaryStore);
  $: timeRangeSummaryError = error as HTTPError;
  // The timeRangeSummary is null when there are 0 rows of data
  // Notably, this happens when a security policy fully restricts a user from reading any data
  $: timeRangeSummaryIsNull =
    data &&
    data.timeRangeSummary?.min === null &&
    data.timeRangeSummary?.max === null;

  $: if (metricsViewName === $ctxName && $metricsViewSchema?.data?.schema) {
    // Make sure we use the correct sync instance for the current metrics view
    unsubscribe?.();
    unsubscribe = useDashboardUrlSync(ctx, $metricsViewSchema?.data?.schema);
  }

  onDestroy(() => {
    unsubscribe?.();
  });
</script>

{#if schemaError}
  <ErrorPage
    statusCode={schemaError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the schema for this dashboard."
    detail={schemaError?.response?.data?.message}
  />
{:else if timeRangeSummaryError}
  <ErrorPage
    statusCode={timeRangeSummaryError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the time range for this dashboard."
    detail={timeRangeSummaryError?.response?.data?.message}
  />
{:else if timeRangeSummaryIsNull}
  <ErrorPage
    header="Error loading dashboard"
    body="This dashboard currently has no data to display. This may be due to access permissions."
  />
{:else if $dashboardStore}
  <slot />
{/if}
