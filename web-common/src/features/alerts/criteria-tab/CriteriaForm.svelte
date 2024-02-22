<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let formState: any; // svelte-forms-lib's FormState
  export let index: number;

  const { form, errors } = formState;

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
    options={CriteriaOperationOptions}
    placeholder="Operator"
  />
  <Select
    id="compareWith"
    label=""
    options={[{ value: "value" }, { value: "measure" }]}
    placeholder="compare with"
    value={"value"}
  />
  <InputV2
    bind:value={$form["criteria"][index]["value"]}
    error={$errors["criteria"][index]["value"]}
    id="value"
    placeholder={"0"}
  />
</div>
