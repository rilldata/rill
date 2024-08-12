<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { parseCriteriaError } from "@rilldata/web-common/features/alerts/criteria-tab/parseCriteriaError";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import {
    MeasureFilterBaseTypeOptions,
    MeasureFilterComparisonTypeOptions,
    MeasureFilterPercentOfTotalOption,
    MeasureFilterType,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
  import { createForm } from "svelte-forms-lib";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;
  export let index: number;

  const { form, errors, validateField } = formState;

  $: metricsView = useMetricsView(
    $runtime.instanceId,
    $form["metricsViewName"],
  );

  $: measure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["measure"],
  );
  $: measureOptions = [
    {
      value: $form["measure"],
      label: measure?.label?.length ? measure.label : measure?.expression,
    },
  ];
  $: selectedMeasure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["criteria"][index].measure,
  );

  $: hasComparison =
    $form.comparisonTimeRange?.isoDuration ||
    $form.comparisonTimeRange?.isoOffset;
  $: comparisonLabel = $form.comparisonTimeRange
    ? getComparisonLabel($form.comparisonTimeRange).toLowerCase()
    : "";
  $: typeOptions = [
    ...(hasComparison
      ? MeasureFilterComparisonTypeOptions.map((o) => {
          if (
            o.value !== MeasureFilterType.AbsoluteChange &&
            o.value !== MeasureFilterType.PercentChange
          )
            return o;
          return {
            ...o,
            label: `${o.label} ${comparisonLabel}`,
          };
        })
      : MeasureFilterBaseTypeOptions),
    ...(selectedMeasure?.validPercentOfTotal
      ? [MeasureFilterPercentOfTotalOption]
      : []),
  ];

  // Debounce the update of value. This avoids constant refetches
  let value: string = $form["criteria"][index].value1;
  const valueUpdater = debounce(() => {
    if ($form["criteria"][index]) $form["criteria"][index].value1 = value;
    void validateField("criteria");
  }, 500);

  $: groupErr = parseCriteriaError($errors["criteria"], index);
</script>

<div class="flex flex-row gap-2">
  <Select
    bind:value={$form["criteria"][index].measure}
    id="field"
    label=""
    options={measureOptions}
    placeholder="Measure"
    className="w-[160px]"
  />
  <Select
    bind:value={$form["criteria"][index].type}
    id="type"
    label=""
    options={typeOptions}
    placeholder="type"
    className="w-[256px]"
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    options={CriteriaOperationOptions}
    placeholder="Operator"
    className="w-[70px]"
  />
  <!-- Error is not returned as an object for criteria[index]. We instead have parsed groupErr -->
  <Input
    alwaysShowError
    bind:value
    id="value"
    onInput={valueUpdater}
    placeholder={"0"}
  />
</div>
{#if groupErr}
  <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
    {groupErr}
  </div>
{/if}
