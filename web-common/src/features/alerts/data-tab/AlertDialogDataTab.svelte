<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import { createForm } from "svelte-forms-lib";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import FilterChipsReadOnly from "../../dashboards/filters/FilterChipsReadOnly.svelte";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;

  const { form } = formState;

  $: ({ instanceId } = $runtime);

  $: metricsViewName = $form["metricsViewName"]; // memoise to avoid rerenders
  $: metricsView = useMetricsViewValidSpec(instanceId, metricsViewName);

  $: measureOptions =
    $metricsView.data?.measures
      ?.filter(
        (m) =>
          !m.window &&
          m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
      )
      .map((m) => ({
        value: m.name as string,
        label: m.displayName?.length
          ? m.displayName
          : (m.expression ?? (m.name as string)),
      })) ?? [];
  $: dimensionOptions = [
    {
      value: "",
      label: "None",
    },
    ...($metricsView.data?.dimensions?.map((d) => ({
      value: d.name as string,
      label: d.displayName?.length
        ? d.displayName
        : (d.expression ?? (d.name as string)),
    })) ?? []),
  ];
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description="These are inherited from the underlying dashboard view."
    title="Filters"
  >
    <FilterChipsReadOnly
      dimensionThresholdFilters={$form["dimensionThresholdFilters"]}
      filters={$form["whereFilter"]}
      dimensionsWithInlistFilter={$form["dimensionsWithInlistFilter"]}
      exploreName={$form["exploreName"]}
      timeRange={$form["timeRange"]}
      comparisonTimeRange={$form["comparisonTimeRange"]}
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
  <FormSection
    title="Data preview"
    description="Here’s a look at the data you’ve selected above."
  >
    <DataPreview formValues={$form} />
  </FormSection>
</div>
