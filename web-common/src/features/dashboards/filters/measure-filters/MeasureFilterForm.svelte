<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  import PinButton from "../PinButton.svelte";
  import RequiredButton from "../RequiredButton.svelte";

  export let dimensionName: string;
  export let name: string;
  export let label: string;
  export let open: boolean;
  export let filter: MeasureFilterEntry | undefined = undefined;
  export let onApply: (params: {
    dimension: string;
    oldDimension: string;
    filter: MeasureFilterEntry;
  }) => void;
  export let allDimensions: MetricsViewSpecDimension[];
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let pinned = false;
  export let showPinControl = false;
  export let required = false;
  export let showRequiredControl = false;

  const initialValues = {
    dimension: dimensionName,
    operation: filter?.operation ?? MeasureFilterOperationOptions[0].value,
    value1: filter?.value1 ?? "",
    value2: filter?.value2 ?? "",
  };

  const validationSchema = object().shape({
    dimension: string().required(m.common_required()),
    operation: mixed<MeasureFilterOperation>()
      .oneOf(Object.values(MeasureFilterOperation))
      .required(m.common_required()),
    value1: string()
      .required(m.common_required())
      .test("is-numeric", m.common_must_be_number(), (value) => {
        return !isNaN(Number(value)) && value.trim() !== "";
      }),
    value2: string().when("operation", {
      is: (val: MeasureFilterOperation) => expressionIsBetween(val),
      then: (schema) =>
        schema
          .required(m.common_required())
          .test("is-numeric", m.common_must_be_number(), (value) => {
            return !isNaN(Number(value)) && value.trim() !== "";
          }),
      otherwise: (schema) => schema.optional(),
    }),
  });

  const { form, formId, errors, submit, enhance } = superForm(
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
  onkeydown={(e) => {
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
  preventScroll
  id="measure-filter-popover"
>
  {#if showPinControl || showRequiredControl}
    <div
      class="flex flex-row items-center justify-between mb-2 pointer-events-auto"
    >
      <b>{label}</b>

      <div class="flex flex-row items-center gap-x-1">
        {#if showRequiredControl}
          <RequiredButton
            required={!!required}
            onToggleRequired={() => {
              required = !required;
            }}
          />
        {/if}
        {#if showPinControl}
          <PinButton
            pinned={!!pinned}
            onTogglePin={() => {
              pinned = !pinned;
            }}
          />
        {/if}
      </div>
    </div>
  {/if}
  <form
    use:enhance
    autocomplete="off"
    class="flex flex-col gap-y-3"
    id={$formId}
  >
    <Select
      bind:value={$form["dimension"]}
      id="dimension"
      label={m.measure_filter_by_dimension()}
      options={dimensionOptions}
      placeholder={m.measure_filter_select_dimension()}
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
      label={m.measure_filter_threshold()}
      options={MeasureFilterOperationOptions}
    />
    <Input
      bind:value={$form["value1"]}
      errors={$errors["value1"]}
      id="value1"
      onEnter={submit}
      alwaysShowError
      placeholder={isBetweenExpression
        ? m.measure_filter_lower_value()
        : m.measure_filter_enter_number()}
    />

    {#if isBetweenExpression}
      <Input
        bind:value={$form["value2"]}
        errors={$errors["value2"]}
        id="value2"
        placeholder={m.measure_filter_higher_value()}
        alwaysShowError
        onEnter={submit}
      />
    {/if}

    <Button submitForm type="primary" form={$formId}
      >{m.measure_filter_apply()}</Button
    >
  </form>
</Popover.Content>
