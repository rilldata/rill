<script lang="ts">
  import { page } from "$app/stores";
  import FormSection from "../../components/forms/FormSection.svelte";
  import InputV2 from "../../components/forms/InputV2.svelte";
  import Select from "../../components/forms/Select.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import Filters from "../dashboards/filters/Filters.svelte";
  import { useDashboard } from "../dashboards/selectors";
  import DataPreview from "./DataPreview.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  $: dashboardName = $page.params.dashboard;
  $: dashboard = useDashboard($runtime.instanceId, dashboardName);

  $: measures = $dashboard.data?.metricsView?.spec?.measures;
  $: measureOptions =
    measures?.map((measure) => ({
      value: measure.name as string,
      label: measure.label,
    })) ?? [];
  $: dimensions = $dashboard.data?.metricsView?.spec?.dimensions;
  $: dimensionOptions =
    dimensions?.map((dimension) => ({
      value: dimension.name as string,
      label: dimension.label,
    })) ?? [];

  $: console.log($form);
</script>

<div class="flex flex-col gap-y-5">
  <FormSection title="Alert name">
    <InputV2
      on:change={handleChange}
      value={$form["name"]}
      error={$errors["name"]}
      id="name"
      placeholder="My alert"
    />
  </FormSection>
  <FormSection
    title="Filters"
    description="These are inherited from the underlying dashboard view."
  >
    <!-- TODO: make these filters read-only -->
    <Filters />
  </FormSection>
  <FormSection
    title="Alert data"
    description="Select the measures you want to monitor."
  >
    {#if measures}
      <Select
        bind:value={$form["measure"]}
        id="measure"
        label="Choose a measure"
        options={measureOptions}
      />
    {/if}
    {#if dimensions}
      <Select
        bind:value={$form["splitByDimension"]}
        id="splitByDimension"
        label="Split by dimension"
        options={dimensionOptions}
      />
    {/if}
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
      metricsView={dashboardName}
      measure={$form["measure"]}
      dimension={$form["splitByDimension"]}
    />
  </FormSection>
</div>
