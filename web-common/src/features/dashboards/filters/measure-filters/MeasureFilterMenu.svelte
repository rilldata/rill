<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { Menu } from "@rilldata/web-common/components/menu";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    createBetweenExpression,
    createBinaryExpression,
  } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import {
    type V1Expression,
    V1Operation,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let dimensionName: string;
  export let name: string;
  export let expr: V1Expression | undefined;

  const dispatch = createEventDispatcher();

  const {
    selectors: {
      dimensions: { allDimensions },
    },
  } = getStateManagers();

  $: dimensionOptions =
    $allDimensions?.map((d) => ({
      value: d.name as string,
      label: getDimensionDisplayName(d),
    })) ?? [];

  let operation: string = MeasureFilterOptions[0].value;
  let value1;
  let value2;
  if (!expr?.cond?.op) {
    operation = MeasureFilterOptions[0].value;
    value1 = undefined;
    value2 = undefined;
  } else {
    if (expr?.cond?.op === V1Operation.OPERATION_AND) {
      operation = "b";
      value1 = expr?.cond?.exprs?.[0].cond?.exprs?.[1]?.val as string;
      value2 = expr?.cond?.exprs?.[1].cond?.exprs?.[1]?.val as string;
    } else if (expr?.cond?.op === V1Operation.OPERATION_OR) {
      operation = "nb";
      value1 = expr?.cond?.exprs?.[0].cond?.exprs?.[1]?.val as string;
      value2 = expr?.cond?.exprs?.[1].cond?.exprs?.[1]?.val as string;
    } else {
      operation = expr?.cond?.op;
      value1 = expr?.cond?.exprs?.[1]?.val as string;
    }
  }

  const formState = createForm({
    initialValues: {
      dimension: dimensionName,
      operation,
      value1,
      value2,
    },
    validationSchema: yup.object({
      dimension: yup.string().required("Required"),
      operation: yup.string().required("Required"),
      value1: yup.number().required("Required"),
      value2: yup.number(),
    }),
    onSubmit: (values) => {
      let newExpr: V1Expression;
      if (values.operation === "b" || values.operation === "nb") {
        if (values.value2 === undefined) {
          return;
        }
        newExpr = createBetweenExpression(
          name,
          Number(values.value1),
          Number(values.value2),
          values.operation === "nb",
        );
      } else {
        newExpr = createBinaryExpression(
          name,
          values.operation as V1Operation,
          Number(values.value1),
        );
      }
      dispatch("apply", {
        dimension: values.dimension,
        expr: newExpr,
      });
    },
  });

  const { form, errors, handleSubmit } = formState;

  let oprn = $form["operation"];
  $: if ($form["operation"] !== oprn) {
    oprn = $form["operation"];
    handleSubmit(new SubmitEvent(""));
  }

  $: isBetweenExpression = oprn === "b" || oprn === "nb";
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
  <div class="p-4">
    <form
      autocomplete="off"
      class="flex flex-col gap-y-3"
      id="measure"
      on:submit|preventDefault={handleSubmit}
    >
      <Select
        bind:value={$form["dimension"]}
        id="operation"
        label="By Dimension"
        options={dimensionOptions}
      />
      <Select
        bind:value={$form["operation"]}
        id="operation"
        label="Threshold"
        options={MeasureFilterOptions}
      />
      <InputV2
        bind:value={$form["value1"]}
        error={$errors["value1"]}
        id="value1"
        on:change={handleSubmit}
        on:enter-pressed={(e) => {
          // TODO: focus next input if isBetweenExpression
          handleSubmit(e);
        }}
      />
      {#if isBetweenExpression}
        <InputV2
          bind:value={$form["value2"]}
          error={$errors["value2"]}
          id="value2"
          on:change={handleSubmit}
          on:enter-pressed={handleSubmit}
        />
      {/if}
    </form>
  </div>
</Menu>
