<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import FilterChipsReadOnly from "../../dashboards/filters/FilterChipsReadOnly.svelte";
  import { useMetricsView } from "../../dashboards/selectors";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  $: metricsViewName = $form["metricsViewName"]; // memoise to avoid rerenders
  $: metricsView = useMetricsView($runtime.instanceId, metricsViewName);

  $: measureOptions =
    $metricsView.data?.measures?.map((m) => ({
      value: m.name as string,
      label: m.label?.length ? m.label : m.expression,
    })) ?? [];
  $: dimensionOptions = [
    {
      value: "",
      label: "None",
    },
    ...($metricsView.data?.dimensions?.map((d) => ({
      value: d.name as string,
      label: d.label?.length ? d.label : d.expression,
    })) ?? []),
  ];
</script>

<div class="flex flex-col gap-y-3">
  <FormSection title="Alert name">
    <InputV2
      alwaysShowError
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
    <FilterChipsReadOnly
      dimensionThresholdFilters={[]}
      filters={$form["whereFilter"]}
      metricsViewName={$form["metricsViewName"]}
      timeRange={$form["timeRange"]}
    />
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
      optional
      options={dimensionOptions}
      placeholder="Select a dimension"
    />
  </FormSection>
  <FormSection title="Data preview">
    <DataPreview
      measure={$form["measure"]}
      metricsViewName={$form["metricsViewName"]}
      splitByDimension={$form["splitByDimension"]}
      timeRange={$form["timeRange"]}
      whereFilter={$form["whereFilter"]}
    />
  </FormSection>
</div>
