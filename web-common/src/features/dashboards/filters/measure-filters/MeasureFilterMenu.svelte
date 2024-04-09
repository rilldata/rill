<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import {
    MeasureFilterComparisonType,
    MeasureFilterEntry,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import {
    MeasureFilterOperation,
    MeasureFilterOptions,
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
    operation: filter?.operation ?? MeasureFilterOptions[0].value,
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
            comparison: MeasureFilterComparisonType.None,
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
      detach
      id="dimension"
      itemsClass="absolute left-4.5"
      label="By Dimension"
      on:change={handleSubmit}
      options={dimensionOptions}
      placeholder="Select dimension to split by"
    />
    <Select
      bind:value={$form["operation"]}
      detach
      id="operation"
      itemsClass="absolute left-4.5"
      label="Threshold"
      on:change={handleSubmit}
      options={MeasureFilterOptions}
    />
    <InputV2
      bind:value={$form["value1"]}
      error={$errors["value1"]}
      id="value1"
      on:change={(e) => {
        handleSubmit(e);
      }}
      on:enter-pressed={handleSubmit}
      placeholder={isBetweenExpression ? "Lower Value" : "Enter a Number"}
    />
    {#if isBetweenExpression}
      <InputV2
        bind:value={$form["value2"]}
        error={$errors["value2"]}
        id="value2"
        placeholder="Higher Value"
        on:change={handleSubmit}
        on:enter-pressed={handleSubmit}
      />
    {/if}
  </form>
</DropdownMenu.Content>
