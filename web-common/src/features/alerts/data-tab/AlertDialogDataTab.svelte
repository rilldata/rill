<script lang="ts">
  import { AlertIntervalOptions } from "@rilldata/web-common/features/alerts/data-tab/intervals";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import FilterChips from "../../dashboards/filters/FilterChips.svelte";
  import AlertDataPreview from "web-common/src/features/alerts/AlertDataPreview.svelte";
  import NoFiltersSelected from "./NoFiltersSelected.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  const {
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

<div class="flex flex-col gap-y-3">
  <FormSection title="Alert name">
    <InputV2
      error={$errors["name"]}
      id="name"
      on:change={handleChange}
      placeholder="My alert"
      value={$form["name"]}
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
      options={measureOptions}
      placeholder="Select a measure"
    />
    <Select
      bind:value={$form["splitByDimension"]}
      id="splitByDimension"
      label="Split by dimension"
      options={dimensionOptions}
      placeholder="Select a dimension"
    />
    <Select
      bind:value={$form["splitByTimeGrain"]}
      id="splitByTimeGrain"
      label="Split By Time Grain"
      options={AlertIntervalOptions}
      placeholder="Select a time grain"
    />
  </FormSection>
  <FormSection title="Data preview">
    <AlertDataPreview
      dimension={$form["splitByDimension"]}
      measure={$form["measure"]}
      splitByTimeGrain={$form["splitByTimeGrain"]}
    />
  </FormSection>
</div>
