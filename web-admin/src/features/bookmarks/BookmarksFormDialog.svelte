<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    type BookmarkEntry,
    type BookmarkFormValues,
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
  import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
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
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import type { Interval } from "luxon";

  const baseFilterState = {
    filters: createAndExpression([]),
    dimensionsWithInlistFilter: [],
    dimensionThresholdFilters: [],
    queryTimeStart: "",
    queryTimeEnd: "",
    displayTimeRange: { expression: "" } as V1TimeRange,
    selectedTimeRange: "",
  };

  export let organization: string;
  export let project: string;
  export let projectId: string;
  export let resource: { name: string; kind: ResourceKind };
  export let bookmark: BookmarkEntry | null = null;
  export let defaultUrlParams: URLSearchParams | undefined = undefined;
  export let showFiltersOnly: boolean = true;
  export let metricsViewNames: string[];
  export let onClose = () => {};

  let filterState = baseFilterState;

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

      const { expr, dimensionsWithInlistFilter } = getFiltersFromText(
        searchParamsObj.get(ExploreStateURLParams.Filters) || "",
      );

      const { dimensionFilters, dimensionThresholdFilters } =
        splitWhereFilter(expr);

      const selectedTimeRange = formatTimeRange(start, end, grain, timeZone);

      return <typeof baseFilterState>{
        dimensionThresholdFilters,
        dimensionsWithInlistFilter,
        filters: dimensionFilters,
        queryTimeStart: start,
        queryTimeEnd: end,
        displayTimeRange: timeRange,
        selectedTimeRange,
      };
    } catch {
      return baseFilterState;
    }
  }

  // Adding it here to get a newline in
  const CategoryTooltip = `Your bookmarks can only be viewed by you.
Managed bookmarks will be available to all viewers of this dashboard.`;

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();

  const { form, errors, handleSubmit, handleReset } =
    createForm<BookmarkFormValues>({
      initialValues: {
        displayName: bookmark?.resource.displayName || "Default Label",
        description: bookmark?.resource.description ?? "",
        shared: bookmark?.resource.shared ? "true" : "false",
        filtersOnly: bookmark?.filtersOnly ?? false,
        absoluteTimeRange: bookmark?.absoluteTimeRange ?? false,
      },
      validationSchema: yup.object({
        displayName: yup.string().required("Required"),
        description: yup.string(),
      }),
      onSubmit: async (values) => {
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
          handleReset();
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
    });
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
      id="create-bookmark-dialog"
      on:submit|preventDefault={() => {
        /* Switch was triggering this causing clicking on them submitting the form */
      }}
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
        <ExploreFilterChipsReadOnly {...filterState} {metricsViewNames} />
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
          <div class="text-gray-500 text-sm">
            {filterState.selectedTimeRange}
          </div>
        </Label>
      </div>
    </form>

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow" />
      <Button onClick={onClose} type="secondary">Cancel</Button>
      <Button onClick={handleSubmit} type="primary">Save</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
