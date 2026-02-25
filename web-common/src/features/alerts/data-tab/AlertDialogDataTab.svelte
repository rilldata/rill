<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import type { SuperForm } from "sveltekit-superforms/client";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let filters: Filters;
  export let timeControls: TimeControls;

  const runtimeClient = useRuntimeClient();

  $: ({ form } = superFormInstance);

  $: metricsViewName = $form["metricsViewName"]; // memoise to avoid rerenders
  $: metricsView = useMetricsViewValidSpec(
    runtimeClient.instanceId,
    metricsViewName,
  );

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
  <FormSection title="Filters">
    <FiltersForm
      instanceId={runtimeClient.instanceId}
      {filters}
      {timeControls}
      maxWidth={750}
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
    <DataPreview formValues={$form} {filters} {timeControls} />
  </FormSection>
</div>
