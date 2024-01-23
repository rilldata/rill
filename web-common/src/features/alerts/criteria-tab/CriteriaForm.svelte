<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
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

  const { form, errors, handleChange } = formState;
</script>

<div class="grid grid-cols-2 flex-wrap gap-2">
  <Select
    bind:value={$form["criteria"][index].field}
    id="field"
    label=""
    options={measureOptions}
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    options={MeasureFilterOptions}
  />
  <Select
    bind:value={$form["criteria"][index].compareWith}
    id="compareWith"
    label=""
    options={[{ value: "value" }, { value: "measure" }]}
  />
  <InputV2
    bind:value={$form["criteria"][index].value}
    error={$errors["value"]}
    id="value"
    on:change={handleChange}
    placeholder={"0"}
  />
</div>
