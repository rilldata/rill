<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import type { SuperForm } from "sveltekit-superforms/client";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import {
    DashboardConfigProvider,
    ExploreDashboardConfigProvider,
  } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let expressionFilterManager: ExpressionFilterManager;
  export let timeFilterManager: TimeFilterManager;

  const runtimeClient = useRuntimeClient();

  $: ({ form } = superFormInstance);

  // memoise to avoid rerenders
  $: metricsViewName = $form["metricsViewName"];
  $: exploreName = $form["exploreName"];
  $: metricsView = useMetricsViewValidSpec(runtimeClient, metricsViewName);

  let dashboardConfigProvider: DashboardConfigProvider;
  $: {
    dashboardConfigProvider?.cleanup?.();
    dashboardConfigProvider = new ExploreDashboardConfigProvider(
      runtimeClient,
      exploreName,
    );
  }

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
    <FiltersForm
      {expressionFilterManager}
      {timeFilterManager}
      {dashboardConfigProvider}
      maxWidth={750}
    />
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
    <DataPreview
      formValues={$form}
      {expressionFilterManager}
      {timeFilterManager}
    />
  </FormSection>
</div>
