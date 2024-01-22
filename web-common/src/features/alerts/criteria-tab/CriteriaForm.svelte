<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createBinaryExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import {
    type V1Expression,
    V1Operation,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let expr: V1Expression;

  const dispatch = createEventDispatcher();

  const {
    selectors: {
      measures: { allMeasures },
    },
  } = getStateManagers();

  const operation =
    expr.cond?.op === V1Operation.OPERATION_UNSPECIFIED
      ? undefined
      : expr.cond?.op;

  const formState = createForm({
    initialValues: {
      field: expr.cond?.exprs?.[0].ident ?? "",
      operation,
      compareWith: "value",
      value: expr.cond?.exprs?.[1].val,
    },
    validationSchema: yup.object({
      field: yup.string().required("Required"),
      operation: yup.string().required("Required"),
      compareWith: yup.string().required("Required"),
      value: yup.number().required("Required"),
    }),
    onSubmit: (values) => {
      dispatch(
        "update",
        createBinaryExpression(
          values.field,
          values.operation as V1Operation,
          Number(values.value),
        ),
      );
    },
  });

  const { form, errors, handleSubmit } = formState;

  $: measureOptions =
    $allMeasures?.map((m) => ({
      value: m.name as string,
      label: m.label?.length ? m.label : m.expression,
    })) ?? [];

  $: console.log($errors);
</script>

<form
  autocomplete="off"
  class="flex flex-col gap-y-3"
  id="measure"
  on:submit|preventDefault={handleSubmit}
>
  <div class="grid grid-cols-2 flex-wrap gap-2">
    <Select
      bind:value={$form["field"]}
      id="field"
      label=""
      options={measureOptions}
    />
    <Select
      bind:value={$form["operation"]}
      id="operation"
      label=""
      options={MeasureFilterOptions}
    />
    <Select
      bind:value={$form["compareWith"]}
      id="compareWith"
      label=""
      options={[{ value: "value" }, { value: "measure" }]}
    />
    <InputV2
      bind:value={$form["value"]}
      error={$errors["value"]}
      id="value"
      on:change={handleSubmit}
      placeholder={"0"}
    />
  </div>
</form>
