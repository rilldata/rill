<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@rilldata/svelte-query";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import { getPrettySelectedTimeRange } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { getProjectPermissions } from "@rilldata/web-admin/features/projects/selectors";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { createForm } from "svelte-forms-lib";

  export let metricsViewName: string;
  export let formState: ReturnType<typeof createForm<BookmarkFormValues>>;
  export let editForm: boolean;

  const queryClient = useQueryClient();
  $: dashboardStore = useDashboardStore(metricsViewName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $dashboardStore.selectedTimeRange?.name,
    start: $dashboardStore.selectedTimeRange?.start?.toISOString() ?? "",
    end: $dashboardStore.selectedTimeRange?.end?.toISOString() ?? "",
  };

  $: projectPermissions = getProjectPermissions(
    $page.params.organization,
    $page.params.project,
  );
  $: manageProject = $projectPermissions.data?.manageProject;

  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    $runtime?.instanceId,
    metricsViewName,
  );

  const { form, errors } = formState;
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
    label={editForm ? "Rename" : "Name"}
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
  <InputV2
    bind:value={$form["description"]}
    error={$errors["description"]}
    id="description"
    label="Description"
    optional
  />
  {#if !editForm && manageProject}
    <Select
      bind:value={$form["shared"]}
      id="shared"
      label="Category"
      options={[
        { value: "false", label: "Your bookmarks" },
        { value: "true", label: "Default bookmarks" },
      ]}
    />
  {/if}
  <div class="flex items-center space-x-2">
    <Switch bind:checked={$form["filtersOnly"]} id="filtersOnly" />
    <Label for="filtersOnly">Save filters only</Label>
  </div>
  <div class="flex items-center space-x-2">
    <Switch bind:checked={$form["absoluteTimeRange"]} id="absoluteTimeRange" />
    <Label class="flex flex-col" for="absoluteTimeRange">
      <div class="text-left text-sm font-medium">Absolute time range</div>
      <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
    </Label>
  </div>
</form>
