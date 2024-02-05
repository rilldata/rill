<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import Filters from "../../dashboards/filters/Filters.svelte";
  import DataPreview from "./../DataPreview.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures },
      dimensions: { allDimensions },
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

  $: console.log($form);
</script>

<div class="flex flex-col gap-y-5">
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
    description="These are inherited from the underlying dashboard view."
    title="Filters"
  >
    <!-- TODO: make these filters read-only -->
    <Filters readonly />
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
