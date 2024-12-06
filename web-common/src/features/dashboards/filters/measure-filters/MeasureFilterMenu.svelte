<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import {
    MeasureFilterOperation,
    MeasureFilterOperationOptions,
    MeasureFilterType,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let dimensionName: string;
  export let name: string;
  export let filter: MeasureFilterEntry | undefined;
  export let open: boolean;

  const dispatch = createEventDispatcher();

  const {
    selectors: {
      dimensions: { allDimensions },
    },
  } = getStateManagers();

  const initialValues = {
    dimension: dimensionName,
    operation: filter?.operation ?? MeasureFilterOperationOptions[0].value,
    value1: filter?.value1 ?? "",
    value2: filter?.value2 ?? "",
  };

  const validationSchema = yup.object().shape({
    dimension: yup.string().required("Required"),
    operation: yup.string().required("Required"),
    value1: yup.number().required("Required"),
    value2: yup.number().when("operation", {
      is: (val: MeasureFilterOperation) => expressionIsBetween(val),
      then: (schema) => schema.required("Required"),
      otherwise: (schema) => schema.optional(),
    }),
  });

  const { form, errors, handleSubmit, updateField, updateValidateField } =
    createForm({
      initialValues,
      validationSchema,
      onSubmit: (values) => {
        lastValidState = { ...values };

        dispatch("apply", {
          dimension: values.dimension,
          oldDimension: dimensionName,
          filter: <MeasureFilterEntry>{
            measure: name,
            operation: values.operation,
            type: MeasureFilterType.Value,
            value1: values.value1,
            value2: values.value2,
          },
        });
      },
    });

  let lastValidState = { ...initialValues };

  $: if (!open) {
    Object.entries(lastValidState).forEach(
      ([key, value]: [keyof typeof initialValues, string]) => {
        updateValidateField(key, value);
      },
    );
  }

  $: selectedOperation = $form.operation;

  $: isBetweenExpression = expressionIsBetween(selectedOperation);

  $: dimensionOptions =
    $allDimensions?.map((d) => ({
      value: d.name as string,
      label: getDimensionDisplayName(d),
    })) ?? [];

  $: if (!isBetweenExpression) updateField("value2", undefined);

  function expressionIsBetween(op: MeasureFilterOperation) {
    return (
      selectedOperation === MeasureFilterOperation.Between ||
      op === MeasureFilterOperation.NotBetween
    );
  }
</script>

<DropdownMenu.Content align="start" class="p-2 px-3 w-[250px]">
  <form
    autocomplete="off"
    class="flex flex-col gap-y-3"
    id="measure"
    on:submit|preventDefault={handleSubmit}
  >
    <Select
      bind:value={$form["dimension"]}
      id="dimension"
      label="By Dimension"
      on:change={handleSubmit}
      options={dimensionOptions}
      placeholder="Select dimension to split by"
    />
    <Select
      bind:value={$form["operation"]}
      id="operation"
      label="Threshold"
      on:change={handleSubmit}
      options={MeasureFilterOperationOptions}
    />
    <Input
      bind:value={$form["value1"]}
      errors={$errors["value1"]}
      id="value1"
      onBlur={handleSubmit}
      onEnter={handleSubmit}
      placeholder={isBetweenExpression ? "Lower Value" : "Enter a Number"}
    />
    {#if isBetweenExpression}
      <Input
        bind:value={$form["value2"]}
        errors={$errors["value2"]}
        id="value2"
        placeholder="Higher Value"
        onBlur={handleSubmit}
        onEnter={handleSubmit}
      />
    {/if}
  </form>
</DropdownMenu.Content>
