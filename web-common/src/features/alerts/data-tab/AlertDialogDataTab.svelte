<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import {
    AIConditionType,
    CriteriaConditionType,
  } from "@rilldata/web-common/features/alerts/ai-alert-utils";
  import DataPreview from "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte";
  import type {
    AlertConditionType,
    AlertFormValues,
  } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
  import type { SuperForm } from "sveltekit-superforms/client";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ephemeralMeasureToSpecMeasure } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { useMetricsViewValidSpec } from "../../dashboards/selectors";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let filters: ExpressionFilterManager;
  export let timeControls: TimeControls;
  // Switches between criteria and an AI condition. The form owns it, since what it remembers must outlive this tab.
  export let setConditionType: (conditionType: AlertConditionType) => void;

  const runtimeClient = useRuntimeClient();

  $: ({ form } = superFormInstance);

  // An alert fires on criteria over a measure, or when the AI analyst judges that a Condition written as a sentence holds.
  // The mode is chosen here because it changes which of the fields below are needed.
  const conditionTypes: AlertConditionType[] = [
    CriteriaConditionType,
    AIConditionType,
  ];
  $: conditionTypeLabels = [
    m.alert_form_condition_type_criteria(),
    m.alert_form_condition_type_ai(),
  ];

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
  $: isAIAlert = $form["conditionType"] === AIConditionType;
  $: optionalMeasureOptions = [
    { value: "", label: m.alert_form_data_none() },
    ...measureOptions,
  ];
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
  <FieldSwitcher
    fields={conditionTypeLabels}
    selected={conditionTypes.indexOf($form.conditionType)}
    onClick={(i) => setConditionType(conditionTypes[i])}
  />
  <FormSection title={m.alert_form_data_filters()}>
    <FiltersForm
      {filters}
      {metricsViewName}
      {exploreName}
      {timeControls}
      maxWidth={750}
    />
  </FormSection>
  <!-- For an AI alert, the measure and the dimension only narrow what the AI looks at, so both are optional. -->
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
      options={isAIAlert ? optionalMeasureOptions : measureOptions}
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
