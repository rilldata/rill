<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    type BookmarkEntry,
    formatTimeRange,
    getBookmarkData,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import { deriveInterval } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    V1TimeGrain,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { InfoIcon } from "lucide-svelte";
  import type { Interval } from "luxon";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import CanvasFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/CanvasFilterChipsReadOnly.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, boolean } from "yup";

  export let organization: string;
  export let project: string;
  export let projectId: string;
  export let resource: { name: string; kind: ResourceKind };
  export let bookmark: BookmarkEntry | null = null;
  export let defaultUrlParams: URLSearchParams | undefined = undefined;
  export let showFiltersOnly: boolean = true;
  export let metricsViewNames: string[];
  export let onClose = () => {};

  let filterState: undefined | Awaited<ReturnType<typeof processUrl>> =
    undefined;

  $: ({ instanceId } = $runtime);

  $: ({ name: resourceName, kind: resourceKind } = resource);

  $: ({ url } = $page);
  $: curUrlParams = url.searchParams;

  $: bookmarkUrl = bookmark?.url || $page.url.searchParams.toString();

  $: processUrl(bookmarkUrl)
    .then((state) => {
      filterState = state;
    })
    .catch(console.error);

  async function processUrl(searchParams: string) {
    const searchParamsObj = new URLSearchParams(searchParams);
    const rangeExpression = searchParamsObj.get(
      ExploreStateURLParams.TimeRange,
    );
    const timeRange = <V1TimeRange>{
      expression: rangeExpression || "",
    };

    const timeZone =
      searchParamsObj.get(ExploreStateURLParams.TimeZone) || "UTC";

    try {
      const promises = metricsViewNames.map((mvName) =>
        deriveInterval(timeRange.expression || "", mvName, timeZone),
      );

      const intervals = await Promise.all(promises);
      let intervalWithLatestEndPoint:
        | {
            interval: Interval;
            grain?: V1TimeGrain | undefined;
            error?: string;
          }
        | undefined;
      intervals.forEach((response) => {
        if (
          !intervalWithLatestEndPoint ||
          (response.interval.end && intervalWithLatestEndPoint.interval.end
            ? response.interval.end > intervalWithLatestEndPoint.interval.end
            : false)
        ) {
          intervalWithLatestEndPoint = response;
        }
      });

      const start = intervalWithLatestEndPoint.interval.start.toISO();
      const end = intervalWithLatestEndPoint.interval.end.toISO();

      const grain =
        (searchParamsObj.get(ExploreStateURLParams.TimeGrain) as V1TimeGrain) ||
        intervalWithLatestEndPoint.grain ||
        V1TimeGrain.TIME_GRAIN_MINUTE;

      const selectedTimeRange = formatTimeRange(start, end, grain, timeZone);

      if (resource.kind === ResourceKind.Canvas) {
        const uiFilters = getCanvasStore(
          resourceName,
          instanceId,
        ).canvasEntity.filterManager.getUIFiltersFromString(searchParams);

        return {
          uiFilters,
          queryTimeStart: start,
          queryTimeEnd: end,
          displayTimeRange: timeRange,
          selectedTimeRange,
        };
      }

      const { expr, dimensionsWithInlistFilter } = getFiltersFromText(
        searchParamsObj.get(ExploreStateURLParams.Filters) || "",
      );

      const { dimensionFilters, dimensionThresholdFilters } =
        splitWhereFilter(expr);

      return {
        dimensionThresholdFilters,
        dimensionsWithInlistFilter,
        filters: dimensionFilters,
        queryTimeStart: start,
        queryTimeEnd: end,
        displayTimeRange: timeRange,
        selectedTimeRange,
      };
    } catch {
      return undefined;
    }
  }

  // Adding it here to get a newline in
  const CategoryTooltip = `Your bookmarks can only be viewed by you.
Managed bookmarks will be available to all viewers of this dashboard.`;

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const initialValues = {
    displayName: bookmark?.resource.displayName || "Default Label",
    description: bookmark?.resource.description ?? "",
    shared: bookmark?.resource.shared ? "true" : "false",
    filtersOnly: bookmark?.filtersOnly ?? false,
    absoluteTimeRange: bookmark?.absoluteTimeRange ?? false,
  };

  const schema = yup(
    object({
      displayName: string().required("Required"),
      description: string(),
      shared: string(),
      filtersOnly: boolean(),
      absoluteTimeRange: boolean(),
    }),
  );

  const formId = "create-bookmark-dialog";

  const { form, errors, submit, reset, enhance } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const bookmarkData = getBookmarkData({
          curUrlParams,
          defaultUrlParams,
          filtersOnly: values.filtersOnly,
          absoluteTimeRange: values.absoluteTimeRange,
        });

        if (bookmark) {
          await $bookmarkUpdater.mutateAsync({
            data: {
              bookmarkId: bookmark.resource.id,
              displayName: values.displayName,
              description: values.description,
              shared: values.shared === "true",
              urlSearch: bookmarkData,
            },
          });
        } else {
          await $bookmarkCreator.mutateAsync({
            data: {
              displayName: values.displayName,
              description: values.description,
              projectId,
              resourceKind,
              resourceName,
              shared: values.shared === "true",
              urlSearch: bookmarkData,
            },
          });
          reset();
        }
        onClose();

        await queryClient.refetchQueries({
          queryKey: getAdminServiceListBookmarksQueryKey({
            projectId,
            resourceKind,
            resourceName,
          }),
        });
        eventBus.emit("notification", {
          message: bookmark ? "Bookmark updated" : "Bookmark created",
        });
      },
    },
  );
</script>

<Dialog.Root
  open
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>
        {bookmark ? "Edit bookmark" : "Bookmark current view"}
      </Dialog.Title>
    </Dialog.Header>

    <form
      class="flex flex-col gap-4 z-50"
      id={formId}
      use:enhance
      on:submit|preventDefault={submit}
    >
      <Input
        bind:value={$form["displayName"]}
        errors={$errors["displayName"]}
        id="displayName"
        label="Label"
      />
      <Input
        bind:value={$form["description"]}
        errors={$errors["description"]}
        id="description"
        label="Description"
        optional
      />
      <div class="flex flex-col gap-y-2">
        <Label class="flex flex-col gap-y-1 text-sm">
          <div class="text-gray-800 font-medium">Filters</div>
          <div class="text-gray-500">
            Inherited from underlying dashboard view.
          </div>
        </Label>
        {#if filterState && "uiFilters" in filterState}
          <CanvasFilterChipsReadOnly
            col={false}
            uiFilters={filterState.uiFilters}
            timeRangeString={filterState.displayTimeRange.expression}
            comparisonRange={undefined}
            timeStart={filterState.queryTimeStart}
            timeEnd={filterState.queryTimeEnd}
          />
        {:else if filterState}
          <ExploreFilterChipsReadOnly
            filters={filterState.filters}
            dimensionsWithInlistFilter={filterState.dimensionsWithInlistFilter}
            dimensionThresholdFilters={filterState.dimensionThresholdFilters}
            displayTimeRange={filterState.displayTimeRange}
            queryTimeStart={filterState.queryTimeStart}
            queryTimeEnd={filterState.queryTimeEnd}
            {metricsViewNames}
          />
        {/if}
      </div>
      <ProjectAccessControls {organization} {project}>
        <Select
          bind:value={$form["shared"]}
          id="shared"
          label="Category"
          options={[
            { value: "false", label: "Your bookmarks" },
            { value: "true", label: "Managed bookmarks" },
          ]}
          slot="manage-project"
          tooltip={CategoryTooltip}
        />
      </ProjectAccessControls>
      {#if showFiltersOnly}
        <div class="flex items-center space-x-2">
          <Switch
            bind:checked={$form["filtersOnly"]}
            id="filtersOnly"
            label="Filters only"
          />

          <Label
            class="font-normal flex gap-x-1 items-center"
            for="filtersOnly"
          >
            <span>Save filters only</span>
            <Tooltip distance={8}>
              <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
              <TooltipContent
                class="whitespace-pre-line"
                maxWidth="600px"
                slot="tooltip-content"
              >
                Toggling this on will only save the filter set above, not the
                full dashboard layout and state.
              </TooltipContent>
            </Tooltip>
          </Label>
        </div>
      {/if}
      <div class="flex items-center space-x-2">
        <Switch
          bind:checked={$form["absoluteTimeRange"]}
          id="absoluteTimeRange"
          label="Absolute time range"
        />
        <Label class="flex flex-col font-normal" for="absoluteTimeRange">
          <div class="text-left text-sm flex gap-x-1 items-center">
            <span>Absolute time range</span>
            <Tooltip distance={8}>
              <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
              <TooltipContent
                class="whitespace-pre-line"
                maxWidth="600px"
                slot="tooltip-content"
              >
                The bookmark will use the dashboard's relative time if this
                toggle is off.
              </TooltipContent>
            </Tooltip>
          </div>
          {#if filterState}
            <div class="text-gray-500 text-sm">
              {filterState.selectedTimeRange}
            </div>
          {/if}
        </Label>
      </div>
    </form>

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow" />
      <Button onClick={onClose} type="secondary">Cancel</Button>
      <Button onClick={submit} type="primary" form={formId} submitForm>
        Save
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
