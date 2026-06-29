<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
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
  $: metricsView = useMetricsViewValidSpec(runtimeClient, metricsViewName);

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
      label: m.alert_form_data_none(),
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
  <FormSection title={m.alert_form_data_filters()}>
    <FiltersForm {filters} {timeControls} maxWidth={750} />
  </FormSection>
  <FormSection
    description={m.alert_form_data_measures_desc()}
    title={m.alert_form_data_title()}
  >
    <Select
      bind:value={$form["measure"]}
      id="measure"
      label={m.alert_form_data_measure()}
      options={measureOptions}
      placeholder={m.alert_form_data_measure_placeholder()}
    />
    <Select
      bind:value={$form["splitByDimension"]}
      id="splitByDimension"
      label={m.alert_form_data_split_by()}
      optional
      options={dimensionOptions}
      placeholder={m.alert_form_data_split_placeholder()}
    />
  </FormSection>
  <FormSection
    title={m.alert_form_data_preview()}
    description={m.alert_form_data_preview_desc()}
  >
    <DataPreview formValues={$form} {filters} {timeControls} />
  </FormSection>
</div>
