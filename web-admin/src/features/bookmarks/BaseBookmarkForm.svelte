<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@rilldata/svelte-query";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import { getPrettySelectedTimeRange } from "@rilldata/web-admin/features/bookmarks/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { createForm } from "svelte-forms-lib";

  export let metricsViewName: string;
  export let formState: ReturnType<typeof createForm<BookmarkFormValues>>;

  const queryClient = useQueryClient();
  $: dashboardStore = useDashboardStore(metricsViewName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $dashboardStore.selectedTimeRange?.name,
    start: $dashboardStore.selectedTimeRange?.start?.toISOString() ?? "",
    end: $dashboardStore.selectedTimeRange?.end?.toISOString() ?? "",
  };

  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    $runtime?.instanceId,
    metricsViewName,
  );

  const { form, errors } = formState;

  // Adding it here to get a newline in
  const CategoryTooltip = `Your bookmarks can only be viewed by you.
Managed bookmarks will be available to all viewers of this dashboard.`;
</script>

<form
  class="flex flex-col gap-4 z-50"
  id="create-bookmark-dialog"
  on:submit|preventDefault={() => {
    /* Switch was triggering this causing clicking on them submitting the form */
  }}
>
  <InputV2
    bind:value={$form["displayName"]}
    error={$errors["displayName"]}
    id="displayName"
    label="Name"
  />
  <InputV2
    bind:value={$form["description"]}
    error={$errors["description"]}
    id="description"
    label="Description"
    optional
  />
  <FormSection
    description={"Inherited from underlying dashboard view."}
    padding=""
    title="Filters"
  >
    <FilterChipsReadOnly
      dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
      filters={$dashboardStore.whereFilter}
      {metricsViewName}
      {timeRange}
    />
  </FormSection>
  <ProjectAccessControls
    organization={$page.params.organization}
    project={$page.params.project}
  >
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
    <Switch bind:checked={$form["filtersOnly"]} id="filtersOnly" />
    <Label class="font-normal flex gap-x-1 items-center" for="filtersOnly">
      <span>Save filters only</span>
      <Tooltip distance={8}>
        <InfoCircle />
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
    <Switch bind:checked={$form["absoluteTimeRange"]} id="absoluteTimeRange" />
    <Label class="flex flex-col font-normal" for="absoluteTimeRange">
      <div class="text-left text-sm flex gap-x-1 items-center">
        <span>Absolute time range</span>
        <Tooltip distance={8}>
          <InfoCircle />
          <TooltipContent
            class="whitespace-pre-line"
            maxWidth="600px"
            slot="tooltip-content"
          >
            The bookmark will use the dashboard's relative time if this toggle
            is off.
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
    </Label>
  </div>
</form>
