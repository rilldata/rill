<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AIConditionType,
    CriteriaConditionType,
  } from "@rilldata/web-common/features/alerts/ai-alert-utils";
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import type { SuperForm } from "sveltekit-superforms/client";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Input from "../../../components/forms/Input.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ephemeralMeasureToSpecMeasure } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
  import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let expressionFilterManager: ExpressionFilterManager;
  export let timeFilterManager: TimeFilterManager;
  export let dashboardConfigProvider: DashboardConfigProvider;
  // False for alerts created from an embed (see AlertForm).
  export let aiConditionAllowed: boolean;

  const runtimeClient = useRuntimeClient();

  $: ({ form, errors } = superFormInstance);

  // For an AI alert, the measure, dimension, filters and time range only narrow what the AI looks at.
  // Its condition says what to compare with, so the time filters don't offer a comparison.
  // The analyst only gets the time range's start and end, so they don't offer another time dimension either.
  $: isAIAlert = $form["conditionType"] === AIConditionType;
  const aiTimeFiltersConfig: TimeFiltersConfig = {
    showTimeDimensionSelector: false,
    showFullRange: true,
    showWatermark: true,
    showComparisonSelector: false,
  };

  // memoise to avoid rerenders
  $: metricsViewName = $form["metricsViewName"];
  $: metricsView = useMetricsViewValidSpec(runtimeClient, metricsViewName);

  // Ephemeral measures have no spec entry; synthesize one so they can be
  // selected as the alert's measure.
  // An AI alert's resolver only gets measure names, so the analyst can't resolve an ephemeral measure's formula.
  $: ephemeralMeasures = $form["ephemeralMeasures"] ?? [];
  $: measureOptions = [
    ...($metricsView.data?.measures ?? []),
    ...(isAIAlert ? [] : ephemeralMeasures.map(ephemeralMeasureToSpecMeasure)),
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
  $: noneOption = { value: "", label: m.alert_form_data_none() };
  $: dimensionOptions = [
    noneOption,
    ...($metricsView.data?.dimensions?.map((d) => ({
      value: d.name as string,
      label: d.displayName?.length
        ? d.displayName
        : (d.expression ?? (d.name as string)),
    })) ?? []),
  ];
</script>

<div class="flex flex-col gap-y-3">
  {#if aiConditionAllowed}
    <div class="grid grid-cols-2">
      <Button
        onClick={() => ($form["conditionType"] = CriteriaConditionType)}
        selected={!isAIAlert}
      >
        {m.alert_form_condition_type_criteria()}
      </Button>
      <Button
        onClick={() => {
          $form["conditionType"] = AIConditionType;
          if (ephemeralMeasures.some((d) => d.name === $form["measure"])) {
            $form["measure"] = "";
          }
        }}
        selected={isAIAlert}
      >
        {m.alert_form_condition_type_ai()}
      </Button>
    </div>
  {/if}
  <FormSection title={m.alert_form_data_filters()}>
    <FiltersForm
      {expressionFilterManager}
      {timeFilterManager}
      {dashboardConfigProvider}
      timeFiltersConfig={isAIAlert ? aiTimeFiltersConfig : undefined}
      maxWidth={750}
    />
  </FormSection>
  <FormSection
    description={isAIAlert
      ? m.alert_form_data_ai_desc()
      : m.alert_form_data_measures_desc()}
    title={m.alert_form_data_title()}
  >
    <Select
      bind:value={$form["measure"]}
      id="measure"
      label={m.alert_form_data_measure()}
      optional={isAIAlert}
      options={isAIAlert ? [noneOption, ...measureOptions] : measureOptions}
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
    <!-- An AI alert notifies with the AI's reason rather than matching rows. -->
    {#if $form["splitByDimension"] && !isAIAlert}
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
    <DataPreview
      formValues={$form}
      {expressionFilterManager}
      {timeFilterManager}
    />
  </FormSection>
</div>
