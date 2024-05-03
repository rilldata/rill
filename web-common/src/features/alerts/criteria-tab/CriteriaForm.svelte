<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    CompareWithOptions,
    CriteriaOperationComparisonOptions,
    CriteriaOperationOptions,
  } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { parseCriteriaError } from "@rilldata/web-common/features/alerts/criteria-tab/parseCriteriaError";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
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

  // Debounce the update of value. This avoid constant refetches
  let value: string = $form["criteria"][index].value;
  const valueUpdater = debounce(() => {
    $form["criteria"][index].value = value;
    void validateField("criteria");
  }, 500);

  $: groupErr = parseCriteriaError($errors["criteria"], index);
</script>

<div class="grid grid-cols-2 flex-wrap gap-2">
  <Select
    bind:value={$form["criteria"][index].field}
    id="field"
    label=""
    options={measureOptions}
    placeholder="Measure"
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    options={hasComparison
      ? CriteriaOperationComparisonOptions
      : CriteriaOperationOptions}
    placeholder="Operator"
  />
  <Select
    bind:value={$form["criteria"][index].compareWith}
    id="compareWith"
    label=""
    options={CompareWithOptions}
    placeholder="compare with"
  />
  <InputV2
    alwaysShowError
    bind:value
    error=""
    id="value"
    on:input={valueUpdater}
    placeholder={"0"}
  />
</div>
{#if groupErr}
  <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
    {groupErr}
  </div>
{/if}
