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
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { InfoIcon } from "lucide-svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let metricsViewName: string;
  export let resourceKind: ResourceKind;
  export let resourceName: string;
  export let bookmark: BookmarkEntry | null = null;
  export let defaultUrlParams: URLSearchParams | undefined = undefined;
  export let filtersState: FiltersState;
  export let timeControlState: TimeControlState;
  export let onClose = () => {};

  $: ({
    params: { organization, project },
    url,
  } = $page);
  $: projectId = useProjectId(organization, project);
  $: curUrlParams = url.searchParams;

  $: start = timeControlState?.selectedTimeRange?.start?.toISOString() ?? "";
  $: end = timeControlState?.selectedTimeRange?.end?.toISOString() ?? "";
  $: timeRange = <V1TimeRange>{
    isoDuration: timeControlState.selectedTimeRange?.name,
    start,
    end,
  };

  $: selectedTimeRange = formatTimeRange(
    start,
    end,
    timeControlState.selectedTimeRange?.interval,
    timeControlState?.selectedTimezone,
  );

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
              data: bookmarkData,
            },
          });
        } else {
          await $bookmarkCreator.mutateAsync({
            data: {
              displayName: values.displayName,
              description: values.description,
              projectId: $projectId.data ?? "",
              resourceKind,
              resourceName,
              shared: values.shared === "true",
              data: bookmarkData,
            },
          });
          handleReset();
        }
        onClose();

        await queryClient.refetchQueries({
          queryKey: getAdminServiceListBookmarksQueryKey({
            projectId: $projectId.data ?? "",
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
        <ExploreFilterChipsReadOnly
          dimensionThresholdFilters={filtersState.dimensionThresholdFilters}
          dimensionsWithInlistFilter={filtersState.dimensionsWithInlistFilter}
          filters={filtersState.whereFilter}
          {metricsViewName}
          displayTimeRange={timeRange}
          queryTimeStart={start}
          queryTimeEnd={end}
        />
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
      <div class="flex items-center space-x-2">
        <Switch
          bind:checked={$form["filtersOnly"]}
          id="filtersOnly"
          label="Filters only"
        />
        <Label class="font-normal flex gap-x-1 items-center" for="filtersOnly">
          <span>Save filters only</span>
          <Tooltip distance={8}>
            <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
            <TooltipContent
              class="whitespace-pre-line"
              maxWidth="600px"
              slot="tooltip-content"
            >
              Toggling this on will only save the filter set above, not the full
              dashboard layout and state.
            </TooltipContent>
          </Tooltip>
        </Label>
      </div>
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
          <div class="text-gray-500 text-sm">{selectedTimeRange}</div>
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
