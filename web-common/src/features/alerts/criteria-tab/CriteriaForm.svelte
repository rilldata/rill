<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { parseCriteriaError } from "@rilldata/web-common/features/alerts/criteria-tab/parseCriteriaError";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import {
    MeasureFilterBaseTypeOptions,
    MeasureFilterComparisonTypeOptions,
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

  $: hasComparison =
    $form.comparisonTimeRange?.isoDuration ||
    $form.comparisonTimeRange?.isoOffset;
  $: comparisonLabel = $form.comparisonTimeRange
    ? getComparisonLabel($form.comparisonTimeRange).toLowerCase()
    : "";
  $: typeOptions = hasComparison
    ? MeasureFilterComparisonTypeOptions.map((o) => {
        if (
          o.value !== MeasureFilterType.AbsoluteChange &&
          o.value !== MeasureFilterType.PercentChange
        )
          return o;
        return {
          ...o,
          label: `${o.label} from ${comparisonLabel}`,
        };
      })
    : MeasureFilterBaseTypeOptions;

  // Debounce the update of value. This avoid constant refetches
  let value: string = $form["criteria"][index].value1;
  const valueUpdater = debounce(() => {
    $form["criteria"][index].value1 = value;
    void validateField("criteria");
  }, 500);

  $: groupErr = parseCriteriaError($errors["criteria"], index);
</script>

<div class="grid grid-cols-12 gap-2">
  <Select
    bind:value={$form["criteria"][index].measure}
    id="field"
    label=""
    options={measureOptions}
    placeholder="Measure"
    className="col-span-3"
  />
  <Select
    bind:value={$form["criteria"][index].type}
    id="type"
    label=""
    options={typeOptions}
    placeholder="type"
    className="col-span-5"
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    options={CriteriaOperationOptions}
    placeholder="Operator"
    className="col-span-1"
  />
  <!-- Error is not returned as an object for criteria[index]. We instead have parsed groupErr -->
  <InputV2
    alwaysShowError
    bind:value
    error=""
    id="value"
    on:input={valueUpdater}
    placeholder={"0"}
    className="col-span-3"
  />
</div>
{#if groupErr}
  <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
    {groupErr}
  </div>
{/if}
