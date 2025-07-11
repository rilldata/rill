<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover/";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import type { MetricsViewSpecDimension } from "@rilldata/web-common/runtime-client";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import {
    MeasureFilterOperation,
    MeasureFilterOperationOptions,
    MeasureFilterType,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { string, object, mixed } from "yup";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let dimensionName: string;
  export let name: string;
  export let open: boolean;
  export let filter: MeasureFilterEntry | undefined = undefined;
  export let onApply: (params: {
    dimension: string;
    oldDimension: string;
    filter: MeasureFilterEntry;
  }) => void;
  export let allDimensions: MetricsViewSpecDimension[];
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  const initialValues = {
    dimension: dimensionName,
    operation: filter?.operation ?? MeasureFilterOperationOptions[0].value,
    value1: filter?.value1 ?? "",
    value2: filter?.value2 ?? "",
  };

  const validationSchema = object().shape({
    dimension: string().required("Required"),
    operation: mixed<MeasureFilterOperation>()
      .oneOf(Object.values(MeasureFilterOperation))
      .required("Required"),
    value1: string()
      .required("Required")
      .test("is-numeric", "Value must be a valid number", (value) => {
        return !isNaN(Number(value)) && value.trim() !== "";
      }),
    value2: string().when("operation", {
      is: (val: MeasureFilterOperation) => expressionIsBetween(val),
      then: (schema) =>
        schema
          .required("Required")
          .test("is-numeric", "Value must be a valid number", (value) => {
            return !isNaN(Number(value)) && value.trim() !== "";
          }),
      otherwise: (schema) => schema.optional(),
    }),
  });

  const { form, errors, submit, enhance } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      validators: yup(validationSchema),
      onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        onApply({
          dimension: values.dimension,
          oldDimension: dimensionName,
          filter: {
            measure: name,
            operation: values.operation,
            type: MeasureFilterType.Value,
            value1: values.value1,
            value2: values.value2 ?? "",
          },
        });

        open = false;
      },
      invalidateAll: false,
      resetForm: false,
    },
  );

  $: ({ operation } = $form);

  $: isBetweenExpression = expressionIsBetween(operation);

  $: dimensionOptions =
    allDimensions.map((d) => ({
      value: d.name as string,
      label: getDimensionDisplayName(d),
    })) ?? [];

  function expressionIsBetween(op: MeasureFilterOperation | string) {
    return (
      isMFO(op) &&
      (op === MeasureFilterOperation.Between ||
        op === MeasureFilterOperation.NotBetween)
    );
  }

  function isMFO(value: string): value is MeasureFilterOperation {
    return value in MeasureFilterOperation;
  }
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Enter") {
      submit();
    }
  }}
/>

<Popover.Content
  align="start"
  {side}
  class="p-2 px-3 w-[250px]"
  strategy="fixed"
>
  <form
    use:enhance
    autocomplete="off"
    class="flex flex-col gap-y-3"
    id="measure"
  >
    <Select
      bind:value={$form["dimension"]}
      id="dimension"
      label="By Dimension"
      options={dimensionOptions}
      placeholder="Select dimension to split by"
    />
    <Select
      bind:value={$form["operation"]}
      onChange={(newValue) => {
        if (!expressionIsBetween(newValue)) {
          form.update(
            ($form) => {
              $form.value2 = "";
              return $form;
            },
            {
              taint: false,
            },
          );
        }
      }}
      id="operation"
      label="Threshold"
      options={MeasureFilterOperationOptions}
    />
    <Input
      bind:value={$form["value1"]}
      errors={$errors["value1"]}
      id="value1"
      onEnter={submit}
      alwaysShowError
      placeholder={isBetweenExpression ? "Lower Value" : "Enter a Number"}
    />

    {#if isBetweenExpression}
      <Input
        bind:value={$form["value2"]}
        errors={$errors["value2"]}
        id="value2"
        placeholder="Higher Value"
        alwaysShowError
        onEnter={submit}
      />
    {/if}

    <Button submitForm type="primary" form="measure">Apply</Button>
  </form>
</Popover.Content>
