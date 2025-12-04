<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil } from "lucide-svelte";
  import { DateTime } from "luxon";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import {
    convertDateToMinutes,
    getExploreFields,
    getSanitizedExploreStateParam,
    hasDashboardDimensionThresholdFilter,
    hasDashboardWhereFilter,
  } from "./form-utils";

  const queryClient = useQueryClient();
  const StateManagers = getStateManagers();

  const {
    dashboardStore,
    metricsViewName,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
    validSpecStore,
  } = StateManagers;

  $: ({ organization, project, dashboard } = $page.params);

  const timeControlStore = useTimeControlStore(StateManagers);

  $: isTitleEmpty = $form.title.trim() === "";

  $: exploreFields = getExploreFields(
    $dashboardStore,
    $visibleDimensions,
    $visibleMeasures,
  );

  $: sanitizedState = getSanitizedExploreStateParam(
    $dashboardStore,
    exploreFields,
    $validSpecStore.data?.explore,
  );

  $: hasWhereFilter = hasDashboardWhereFilter($dashboardStore);
  $: hasDimensionThresholdFilter =
    hasDashboardDimensionThresholdFilter($dashboardStore);
  $: hasSomeFilter = hasWhereFilter || hasDimensionThresholdFilter;

  let url: string | null = null;
  let setExpiration = false;
  let apiError: string;
  let popoverOpen = false;
  let copied = false;

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
          const filter = hasSomeFilter
            ? mergeDimensionAndMeasureFilters(
                $dashboardStore.whereFilter,
                $dashboardStore.dimensionThresholdFilters,
              )
            : undefined;

          const { url: _url } = await $issueMagicAuthToken.mutateAsync({
            org: organization,
            project,
            data: {
              resourceType: ResourceKind.Explore as string,
              resourceName: dashboard,
              filter,
              fields: exploreFields,
              ttlMinutes: setExpiration
                ? convertDateToMinutes(values.expiresAt).toString()
                : undefined,
              state: sanitizedState ? sanitizedState : undefined,
              displayName: values.title,
            },
          });

          url = _url;

          void queryClient.invalidateQueries({
            queryKey: getAdminServiceListMagicAuthTokensQueryKey(
              organization,
              project,
            ),
          });
        } catch (error) {
          const typedError = error as HTTPError;
          apiError = typedError.response?.data?.message ?? typedError.message;
        }
      },
      invalidateAll: false,
    },
  );

  function onCopy() {
    copyToClipboard(url, "URL copied to clipboard", false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }

  $: if (setExpiration && $form.expiresAt === null) {
    // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
    $form.expiresAt = DateTime.now().plus({ days: 60 }).toISO();
    popoverOpen = true;
  } else if (!setExpiration) {
    $form.expiresAt = null;
    popoverOpen = false;
  }

  $: ({ length: allErrorsLength } = $allErrors);

  $: maxExpirationDate = DateTime.now().plus({ years: 1 }).startOf("day");
</script>

{#if !url}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <h3 class="text-xs text-gray-800 font-normal">
      Create a shareable public URL for this view.
    </h3>

    {#if !url}
      <div class="flex flex-col gap-y-1 mt-4">
        <Input
          id="name-input"
          bind:value={$form.title}
          placeholder="Label this URL"
        />
      </div>

      <div class="mt-4" class:mb-4={!hasSomeFilter}>
        <div class="flex items-center gap-x-2">
          <Switch small id="has-expiration" bind:checked={setExpiration} />
          <Label class="text-xs" for="has-expiration">Set expiration</Label>
        </div>
        {#if setExpiration}
          <div class="flex items-center gap-x-1 pl-[30px]">
            <label for="expires-at" class="text-slate-500 font-medium">
              Access expires {new Date($form.expiresAt).toLocaleDateString(
                "en-US",
                { year: "numeric", month: "short", day: "numeric" },
              )}
            </label>
            <Popover bind:open={popoverOpen}>
              <PopoverTrigger>
                <IconButton ariaLabel="Edit expiration date">
                  <Pencil size="14px" class="text-primary-600" />
                </IconButton>
              </PopoverTrigger>
              <PopoverContent align="end" class="p-0" strategy="fixed">
                <Calendar
                  selection={DateTime.fromISO($form.expiresAt)}
                  singleDaySelection
                  maxDate={maxExpirationDate}
                  firstVisibleMonth={DateTime.fromISO($form.expiresAt)}
                  onSelectDay={(date) => {
                    $form.expiresAt = date.toISO();
                    popoverOpen = false;
                  }}
                />
              </PopoverContent>
            </Popover>
          </div>
        {/if}
      </div>

      <!-- TODO: revisit when time range lock is implemented -->
      <!-- <div class="mt-4" class:mb-4={!hasWhereFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="lock-time-range" bind:checked={lockTimeRange} />

        <div class="flex flex-row items-center gap-x-1">
          <Label class="text-xs" for="lock-time-range">Lock time range</Label>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-gray-500">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Only data within this range will be visible
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      {#if lockTimeRange}
        <div class="w-full pl-[30px]">
          <label for="lock-time-range" class="text-slate-500 font-medium">
            {#if interval.isValid}
              <RangeDisplay {interval} grain={activeTimeGrain} {abbreviation} />
            {/if}
          </label>
        </div>
      {/if}
    </div> -->

      {#if hasSomeFilter}
        <hr class="mt-4 mb-4" />

        <div class="flex flex-col gap-y-1">
          <p class="text-xs text-gray-800 font-normal">
            The following filters will be locked and hidden:
          </p>
          <div class="flex flex-row gap-1 mt-2">
            <ExploreFilterChipsReadOnly
              metricsViewNames={[$metricsViewName]}
              filters={$dashboardStore.whereFilter}
              dimensionsWithInlistFilter={$dashboardStore.dimensionsWithInlistFilter}
              dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
              queryTimeStart={$timeControlStore.timeStart}
              queryTimeEnd={$timeControlStore.timeEnd}
            />
          </div>
        </div>

        <p class="text-xs text-gray-800 font-normal mt-4 mb-4">
          Measures and dimensions will be limited to current visible set.
        </p>
      {/if}
    {/if}

    <Button
      type="primary"
      disabled={$submitting || isTitleEmpty}
      form={formId}
      submitForm
    >
      Create
    </Button>

    {#if allErrorsLength > 0}
      {#each $allErrors as error (error.path)}
        <div class="text-red-500 mt-1">{error.messages}</div>
      {/each}
    {:else if apiError}
      <div class="text-red-500 mt-1">{apiError}</div>
    {/if}
  </form>
{:else}
  <div class="flex flex-col gap-y-4">
    <h3>Success! A public URL has been created.</h3>
    <Button
      type="secondary"
      onClick={onCopy}
      dataAttributes={{ "data-public-url": url }}
    >
      {#if copied}
        <Check size="16px" />
        Copied URL
      {:else}
        Copy Public URL
      {/if}
    </Button>
  </div>
{/if}

<style lang="postcss">
  form {
    @apply flex flex-col;
  }

  h3 {
    @apply font-semibold;
  }
</style>
