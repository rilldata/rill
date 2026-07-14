<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { prettyParamLabel } from "@rilldata/web-common/features/custom-viz/params";
  import {
    MetricsViewSpecDimensionType,
    type V1ComponentParam,
    type V1MetricsViewSpec,
  } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";

  export let param: V1ComponentParam;
  export let value: unknown;
  // Names of the project's metrics views (for metrics_view params).
  export let metricsViewNames: string[] = [];
  // The spec of the metrics view currently bound to the param's metrics_view param
  // (for field-typed params).
  export let metricsViewSpec: V1MetricsViewSpec | undefined = undefined;
  // Context-computed choices (e.g. the order_by param selects among the values
  // currently bound to the component's field params).
  export let valueOptions: { value: string; label: string }[] | undefined =
    undefined;
  export let onChange: (value: unknown) => void;

  $: label = prettyParamLabel(param.name);
  $: id = `param-${param.name}`;

  // Scalar text/number inputs commit as the user types (debounced so the
  // preview doesn't refetch per keystroke). Input.svelte only fires its
  // onChange prop for the select variant, so plain inputs must use oninput.
  let commitTimer: ReturnType<typeof setTimeout> | undefined;
  onDestroy(() => clearTimeout(commitTimer));

  function commitScalarInput(event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    const raw = param.type === "number" ? target.valueAsNumber : target.value;
    clearTimeout(commitTimer);
    commitTimer = setTimeout(() => {
      if (param.type === "number") {
        onChange(
          typeof raw === "number" && !Number.isNaN(raw) ? raw : undefined,
        );
      } else {
        onChange(raw === "" ? undefined : raw);
      }
    }, 300);
  }

  $: fieldOptions = getFieldOptions(param, metricsViewSpec);

  function getFieldOptions(
    param: V1ComponentParam,
    spec: V1MetricsViewSpec | undefined,
  ): { value: string; label: string }[] {
    if (!spec) return [];
    switch (param.type) {
      case "measure":
        return (spec.measures ?? []).map((measure) => ({
          value: measure.name ?? "",
          label: measure.displayName || (measure.name ?? ""),
        }));
      case "dimension":
        return (spec.dimensions ?? [])
          .filter(
            (dimension) =>
              dimension.type !==
              MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
          )
          .map((dimension) => ({
            value: dimension.name ?? "",
            label: dimension.displayName || (dimension.name ?? ""),
          }));
      case "time_dimension": {
        const options = (spec.dimensions ?? [])
          .filter(
            (dimension) =>
              dimension.type ===
              MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
          )
          .map((dimension) => ({
            value: dimension.name ?? "",
            label: dimension.displayName || (dimension.name ?? ""),
          }));
        if (
          spec.timeDimension &&
          !options.some((option) => option.value === spec.timeDimension)
        ) {
          options.unshift({
            value: spec.timeDimension,
            label: spec.timeDimension,
          });
        }
        return options;
      }
      default:
        return [];
    }
  }
</script>

{#if param.type === "metrics_view"}
  <Select
    {id}
    {label}
    size="sm"
    full
    sameWidth
    value={(value as string) ?? ""}
    options={metricsViewNames.map((name) => ({ value: name, label: name }))}
    onChange={(next) => onChange(next)}
  />
{:else if param.type === "measure" || param.type === "dimension" || param.type === "time_dimension"}
  <Select
    {id}
    {label}
    size="sm"
    full
    sameWidth
    value={(value as string) ?? ""}
    options={fieldOptions}
    onChange={(next) => onChange(next)}
  />
{:else if param.type === "boolean"}
  <div class="flex items-center justify-between py-1">
    <InputLabel small {label} {id} />
    <Switch
      small
      checked={Boolean(value ?? param.default ?? false)}
      onCheckedChange={(next) => onChange(next)}
    />
  </div>
{:else if valueOptions?.length}
  <Select
    {id}
    {label}
    size="sm"
    full
    sameWidth
    value={(value as string) ?? ""}
    options={valueOptions}
    onChange={(next) => onChange(next)}
  />
{:else if param.options?.length}
  <Select
    {id}
    {label}
    size="sm"
    full
    sameWidth
    value={(value ?? param.default) as string}
    options={param.options.map((option) => ({
      value: option as string,
      label: String(option),
    }))}
    onChange={(next) => onChange(param.type === "number" ? Number(next) : next)}
  />
{:else}
  <Input
    {label}
    size="sm"
    labelGap={2}
    capitalizeLabel={false}
    textClass="text-sm"
    inputType={param.type === "number" ? "number" : "text"}
    value={value === undefined || value === null ? "" : String(value)}
    oninput={commitScalarInput}
  />
{/if}
