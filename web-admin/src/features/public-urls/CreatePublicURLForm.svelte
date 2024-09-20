<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    convertDateToMinutes,
    getMetricsViewFields,
    getSanitizedDashboardStateParam,
    hasDashboardWhereFilter,
  } from "./form-utils";
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import { Divider } from "@rilldata/web-common/components/menu";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  const queryClient = useQueryClient();
  const StateManagers = getStateManagers();

  const {
    dashboardStore,
    metricsViewName,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  $: ({
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    minTimeGrain,
  } = $timeControlsStore);

  // $: console.log("selectedTimeRange: ", selectedTimeRange);
  // $: console.log("allTimeRange: ", allTimeRange);
  // $: console.log("showTimeComparison: ", showTimeComparison);
  // $: console.log("selectedComparisonTimeRange: ", selectedComparisonTimeRange);
  // $: console.log("minTimeGrain: ", minTimeGrain);

  $: ({ organization, project } = $page.params);

  $: isTitleEmpty = $form.title.trim() === "";

  $: metricsViewFields = getMetricsViewFields(
    $dashboardStore,
    $visibleDimensions,
    $visibleMeasures,
  );

  $: sanitizedState = getSanitizedDashboardStateParam(
    $dashboardStore,
    metricsViewFields,
  );

  let token: string;
  let setExpiration = false;
  let lockTimeRange = false;
  let apiError: string;

  const formId = "create-public-url-form";

  const initialValues = {
    expiresAt: null,
    title: "",
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
    title: string().required("Title is required"),
  });

  const issueMagicAuthToken = createAdminServiceIssueMagicAuthToken();

  const { form, enhance, submit, allErrors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          const { token: _token } = await $issueMagicAuthToken.mutateAsync({
            organization,
            project,
            data: {
              metricsView: $metricsViewName,
              metricsViewFilter: hasWhereFilter
                ? $dashboardStore.whereFilter
                : undefined,
              metricsViewFields,
              ttlMinutes: setExpiration
                ? convertDateToMinutes(values.expiresAt).toString()
                : undefined,
              state: sanitizedState ? sanitizedState : undefined,
              title: values.title,
            },
          });
          token = _token;

          copyToClipboard(
            `${window.location.origin}/${organization}/${project}/-/share/${token}`,
            "URL copied to clipboard",
          );

          void queryClient.invalidateQueries(
            getAdminServiceListMagicAuthTokensQueryKey(organization, project),
          );
        } catch (error) {
          const typedError = error as HTTPError;
          apiError = typedError.response?.data?.message ?? typedError.message;
        }
      },
    },
  );

  $: hasWhereFilter = hasDashboardWhereFilter($dashboardStore);
  // $: console.log("hasWhereFilter", hasWhereFilter);

  $: if (setExpiration && $form.expiresAt === null) {
    // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
    $form.expiresAt = new Date(Date.now() + 60 * 24 * 60 * 60 * 1000)
      .toISOString()
      .slice(0, 10); // ISO string formatted for input[type="date"]
  } else if (!setExpiration) {
    $form.expiresAt = null;
  }

  $: ({ length: allErrorsLength } = $allErrors);

  // Minimum date is tomorrow
  $: minExpirationDate = new Date(Date.now() + 24 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 10);

  function formatDate(date: Date) {
    if (!date) return "-";
    return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    }).format(new Date(date));
  }

  $: abbreviation = getAbbreviationForIANA(
    new Date(),
    $dashboardStore.selectedTimezone,
  );

  $: lockTimeRangeLabel = $dashboardStore.selectedTimeRange
    ? `${formatDate($dashboardStore.selectedTimeRange?.start) ?? ""} - ${formatDate($dashboardStore.selectedTimeRange?.end) ?? ""} ${abbreviation}`
    : `${formatDate(allTimeRange?.start) ?? ""} - ${formatDate(allTimeRange?.end) ?? ""} ${abbreviation}`;

  // $: console.log(
  //   "selectedTimeRange",
  //   formatDate($dashboardStore.selectedTimeRange?.start),
  //   formatDate($dashboardStore.selectedTimeRange?.end),
  // );
  // $: console.log(
  //   "allTimeRange",
  //   formatDate(allTimeRange?.start),
  //   formatDate(allTimeRange?.end),
  // );

  $: {
    const timeRange = $dashboardStore.selectedTimeRange || allTimeRange;
    lockTimeRangeLabel = timeRange
      ? `${formatDate(timeRange.start) ?? ""} - ${formatDate(timeRange.end) ?? ""} ${abbreviation}`
      : "";
    console.log("lockTimeRangeLabel", lockTimeRangeLabel);
  }
</script>

{#if !token}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <div class="flex flex-col gap-y-4">
      <h3 class="text-sm text-gray-800 font-semibold">
        Create a shareable public URL for this view
      </h3>

      <div class="name-input-container">
        <!-- TODO: this was added because we populate the label with the title of the dashboard -->
        <!-- <Label for="name-input" class="text-xs">URL label</Label> -->
        <input
          id="name-input"
          type="text"
          bind:value={$form.title}
          placeholder="Label this URL"
          class="w-full px-3 py-2 border border-gray-300 rounded-md"
        />
      </div>
    </div>

    <div class="mt-4">
      <div class="flex items-center gap-x-2">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration">Set expiration</Label>
      </div>
      {#if setExpiration}
        <div class="flex items-center gap-x-2 pl-[30px]">
          <label for="expires-at" class="text-slate-500 font-medium w-2/3">
            Access expires
          </label>
          <input
            id="expires-at"
            type="date"
            bind:value={$form.expiresAt}
            min={minExpirationDate}
            class="w-1/3"
          />
        </div>
      {/if}
    </div>

    <!-- We currently lock time range no matter what -->
    <!-- Does that mean we need to strip time range from the stored public url state if user decides not to lock time range? -->
    <!-- What does locking time range mean? Does public url user get to edit the time range? -->
    <!-- If locked, is it read only? -->
    <!-- TODO: provide ability to not lock time range -->
    <div class="mt-4" class:mb-4={!hasWhereFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="lock-time-range" bind:checked={lockTimeRange} />

        <div class="flex flex-row items-center gap-x-1">
          <Label class="text-xs" for="lock-time-range">Lock time range</Label>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-gray-500">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Lock time range to prevent the user from changing the time range.
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      {#if lockTimeRange}
        <div class="w-full pl-[30px]">
          <label for="lock-time-range" class="text-slate-500 font-medium">
            {lockTimeRangeLabel}
          </label>
        </div>
      {/if}
    </div>

    <!-- NOTE: Measures and dimensions will be limited to current visible set. -->
    <!-- NOTE: Filters will be locked and hidden. -->
    {#if hasWhereFilter}
      <Divider marginTop={4} marginBottom={4} />

      <div class="flex flex-col gap-y-1">
        <p class="text-xs text-gray-800 font-normal">
          The following filters will be locked and hidden:
        </p>
        <div class="flex flex-row gap-1 mt-2">
          <!-- TODO: Why isn't MeasureFilter showing up? -->
          <FilterChipsReadOnly
            metricsViewName={$metricsViewName}
            filters={$dashboardStore.whereFilter}
            dimensionThresholdFilters={[]}
            timeRange={undefined}
            comparisonTimeRange={undefined}
          />
        </div>
      </div>

      <p class="text-xs text-gray-800 font-normal mt-4 mb-4">
        Measures and dimensions will be limited to current visible set.
      </p>
    {/if}

    <Button
      type="primary"
      disabled={$submitting || isTitleEmpty}
      form={formId}
      submitForm
    >
      Create and copy URL
    </Button>

    {#if allErrorsLength > 0}
      {#each $allErrors as error (error.path)}
        <div class="text-red-500">{error.messages}</div>
      {/each}
    {:else if apiError}
      <div class="text-red-500">{apiError}</div>
    {/if}
  </form>
{:else}
  <!-- A successful form submission will automatically copy the link to the clipboard -->
  <div class="flex flex-col gap-y-2">
    <h3>Success! URL copied to clipboard.</h3>
  </div>
{/if}

<style lang="postcss">
  form {
    @apply flex flex-col;
  }

  h3 {
    @apply font-semibold;
  }

  .name-input-container {
    @apply flex flex-col gap-y-1;
  }

  input {
    @apply size-full outline-none border-0;
  }

  #name-input {
    @apply flex justify-center items-center overflow-hidden;
    @apply h-8 pl-2 w-full;
    @apply border border-gray-300 rounded-sm;
    @apply text-xs;
  }

  #name-input:focus-within {
    @apply border-primary-500;
  }

  #name-input::placeholder {
    @apply text-xs;
  }
</style>
