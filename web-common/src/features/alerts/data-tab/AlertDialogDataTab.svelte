<script lang="ts">
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import type { SuperForm } from "sveltekit-superforms/client";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Input from "../../../components/forms/Input.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ephemeralMeasureToSpecMeasure } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { MeasureFilterType } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let filters: ExpressionFilterManager;
  export let timeControls: TimeControls;

  const runtimeClient = useRuntimeClient();

  $: ({ form, errors } = superFormInstance);

  // memoise to avoid rerenders
  $: metricsViewName = $form["metricsViewName"];
  $: exploreName = $form["exploreName"];
  $: metricsView = useMetricsViewValidSpec(runtimeClient, metricsViewName);

  // Ephemeral measures have no spec entry; synthesize one so they can be
  // selected as the alert's measure.
  $: measureOptions = [
    ...($metricsView.data?.measures ?? []),
    ...($form["ephemeralMeasures"] ?? []).map(ephemeralMeasureToSpecMeasure),
  ]
    .filter(
      (m) =>
        !m.window &&
        m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
    )
    .map((m) => ({
      value: m.name as string,
      label: m.displayName?.length
        ? m.displayName
        : (m.expression ?? (m.name as string)),
    }));
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
      {filters}
      {metricsViewName}
      {exploreName}
      {timeControls}
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
      onChange={(dimension) => {
        // Percent-of-total criteria compare each dimension value's share of the measure,
        // so they only make sense with a split; drop them when the split is removed.
        if (dimension) return;
        $form["criteria"] = $form["criteria"].filter(
          (c) => c.type !== MeasureFilterType.PercentOfTotal,
        );
      }}
    />
    {#if $form["splitByDimension"]}
      <Input
        bind:value={$form["notifyRowsLimit"]}
        errors={$errors["notifyRowsLimit"]}
        id="notifyRowsLimit"
        label={m.alert_form_data_rows_limit()}
        capitalizeLabel={false}
        hint={m.alert_form_data_rows_limit_hint()}
      />
    {/if}
  </FormSection>
  <FormSection
    title={m.alert_form_data_preview()}
    description={m.alert_form_data_preview_desc()}
  >
    <DataPreview formValues={$form} {filters} {timeControls} />
  </FormSection>
</div>
