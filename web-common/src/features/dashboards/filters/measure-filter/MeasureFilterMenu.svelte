<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filter/measure-filter-options";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    createAndExpression,
    createBetweenExpression,
    createBinaryExpression,
  } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
  import type {
    V1Expression,
    V1Operation,
  } from "@rilldata/web-common/runtime-client";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  const stateManagers = getStateManagers();
  const {
    actions: {
      measuresFilter: { setMeasureFilter },
    },
  } = stateManagers;

  export let name: string;
  export let expr: V1Expression | undefined;

  const formState = createForm({
    initialValues: {
      operation: expr?.cond?.op ?? MeasureFilterOptions[0].value,
      value1: expr?.cond?.exprs?.[1]?.val as string,
      value2: expr?.cond?.exprs?.[2]?.val as string,
    },
    validationSchema: yup.object({
      operation: yup.string().required("Required"),
      value1: yup.number().required("Required"),
      value2: yup.number(),
    }),
    onSubmit: (values) => {
      let newExpr: V1Expression;
      if (values.operation === "b" || values.operation === "nb") {
        newExpr = createBetweenExpression(
          name,
          Number(values.value1),
          Number(values.value2),
          values.operation === "nb"
        );
      } else {
        newExpr = createBinaryExpression(
          name,
          values.operation as V1Operation,
          Number(values.value1)
        );
      }
      setMeasureFilter(name, newExpr);
    },
  });

  const { form, errors, handleSubmit } = formState;

  let oprn = $form["operation"];
  $: if ($form["operation"] !== oprn) {
    oprn = $form["operation"];
    handleSubmit(new SubmitEvent(""));
  }
</script>

<Menu
  focusOnMount={false}
  maxHeight="400px"
  maxWidth="480px"
  minHeight="150px"
  on:click-outside
  on:escape
  paddingBottom={0}
  paddingTop={1}
  rounded={false}
>
  <form
    autocomplete="off"
    class="flex flex-col gap-y-6"
    id="measure"
    on:submit|preventDefault={handleSubmit}
  >
    <Select
      bind:value={$form["operation"]}
      id="operation"
      label="Operation"
      options={MeasureFilterOptions}
    />
    <InputV2
      bind:value={$form["value1"]}
      error={$errors["value1"]}
      on:input={handleSubmit}
    />
    {#if $form["operation"] === "b" || $form["operation"] === "nb"}
      <InputV2
        bind:value={$form["value2"]}
        error={$errors["value2"]}
        on:input={handleSubmit}
      />
    {/if}
  </form>
</Menu>
