<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let formState: any; // svelte-forms-lib's FormState
  export let index: number;

  const {
    selectors: {
      measures: { allMeasures },
    },
  } = getStateManagers();

  $: measureOptions =
    $allMeasures?.map((m) => ({
      value: m.name as string,
      label: m.label?.length ? m.label : m.expression,
    })) ?? [];

  const { form, errors } = formState;
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
    id="value"
    bind:value={$form["criteria"][index]["value"]}
    error={$errors["criteria"][index]["value"]}
    placeholder={"0"}
  />
</div>
