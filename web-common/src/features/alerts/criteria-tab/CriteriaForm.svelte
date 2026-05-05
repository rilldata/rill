<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getTypeOptions } from "@rilldata/web-common/features/alerts/criteria-tab/getTypeOptions";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { parseCriteriaError } from "@rilldata/web-common/features/alerts/criteria-tab/parseCriteriaError";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { SuperForm } from "sveltekit-superforms/client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let timeControls: TimeControls;
  export let index: number;

  const runtimeClient = useRuntimeClient();

  $: ({ form, errors, validate } = superFormInstance);
  $: ({ selectedComparisonTimeRange } = timeControls);

  $: metricsView = useMetricsViewValidSpec(
    runtimeClient,
    $form["metricsViewName"],
  );

  $: measure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["measure"],
  );
  $: measureOptions = [
    {
      value: $form["measure"],
      label: measure?.displayName?.length
        ? measure.displayName
        : (measure?.expression ?? $form["measure"]),
    },
  ];
  $: selectedMeasure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["criteria"][index].measure,
  );

  $: typeOptions = getTypeOptions(
    $form,
    $selectedComparisonTimeRange,
    selectedMeasure,
  );

  // Debounce the update of value. This avoids constant refetches
  let value: string = "0";
  const valueUpdater = debounce(() => {
    if ($form["criteria"][index]) $form["criteria"][index].value1 = value;
    void validate(`criteria[${index}].value1`);
  }, 500);

  // memoize `type` to avoid unnecessary calls to `validateField("criteria")`
  $: type = $form["criteria"][index].type;
  // changing type should re-trigger `criteria` validation,
  // especially when changed to/from a percent type
  $: if (type) void validate(`criteria[${index}].value1`);

  $: groupErr = parseCriteriaError($errors?.criteria?.[index]);

  onMount(() => {
    value = $form["criteria"][index].value1;
  });
</script>

<div class="flex flex-row gap-2" aria-label="criteria-{index}">
  <Select
    bind:value={$form["criteria"][index].measure}
    id="field"
    label=""
    ariaLabel="Criteria measure"
    options={measureOptions}
    placeholder="Measure"
    width={160}
  />
  <Select
    bind:value={$form["criteria"][index].type}
    id="type"
    label=""
    ariaLabel="Criteria type"
    options={typeOptions}
    placeholder="type"
    width={256}
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    ariaLabel="Criteria operator"
    options={CriteriaOperationOptions}
    placeholder="Operator"
    width={70}
  />
  <!-- Error is not returned as an object for criteria[index]. We instead have parsed groupErr -->
  <Input
    alwaysShowError
    bind:value
    id="value"
    title="Criteria value"
    onInput={valueUpdater}
    placeholder={"0"}
    width="auto"
  />
</div>
{#if groupErr}
  <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
    {groupErr}
  </div>
{/if}
