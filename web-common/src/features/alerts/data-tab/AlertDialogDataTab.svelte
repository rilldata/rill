<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import FilterChips from "../../dashboards/filters/FilterChips.svelte";
  import DataPreview from "./../DataPreview.svelte";
  import NoFiltersSelected from "./NoFiltersSelected.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures },
      dimensions: { allDimensions },
      measureFilters: { hasAtLeastOneMeasureFilter },
      dimensionFilters: { hasAtLeastOneDimensionFilter },
    },
  } = getStateManagers();

  $: measureOptions =
    $allMeasures?.map((measure) => ({
      value: measure.name as string,
      label: measure.label,
    })) ?? [];
  $: dimensionOptions =
    $allDimensions?.map((dimension) => ({
      value: dimension.name as string,
      label: dimension.label,
    })) ?? [];

  $: hasAtLeastOneFilter =
    $hasAtLeastOneDimensionFilter || $hasAtLeastOneMeasureFilter;
</script>

<div class="flex flex-col gap-y-5">
  <FormSection title="Alert name">
    <InputV2
      id="name"
      value={$form["name"]}
      error={$errors["name"]}
      placeholder="My alert"
      on:change={handleChange}
    />
  </FormSection>
  <FormSection
    description={hasAtLeastOneFilter
      ? "These are inherited from the underlying dashboard view."
      : ""}
    title="Filters"
  >
    {#if hasAtLeastOneFilter}
      <FilterChips readOnly />
    {:else}
      <NoFiltersSelected />
    {/if}
  </FormSection>
  <FormSection
    description="Select the measures you want to monitor."
    title="Alert data"
  >
    <Select
      bind:value={$form["measure"]}
      id="measure"
      label="Measure"
      placeholder="Select a measure"
      options={measureOptions}
    />
    <Select
      bind:value={$form["splitByDimension"]}
      id="splitByDimension"
      label="Split by dimension"
      placeholder="Select a dimension"
      options={dimensionOptions}
    />
    <!-- TODO -->
    <!-- <Select
      bind:value={$form["forEvery"]}
      id="forEvery"
      label="For every"
      options={["Interval1", "Interval2", "Interval3"].map((timeInterval) => ({
        value: timeInterval,
      }))}
    /> -->
  </FormSection>
  <FormSection title="Data preview">
    <DataPreview
      dimension={$form["splitByDimension"]}
      filter={$dashboardStore.whereFilter}
      measure={$form["measure"]}
      metricsView={$dashboardStore.name}
    />
  </FormSection>
</div>
