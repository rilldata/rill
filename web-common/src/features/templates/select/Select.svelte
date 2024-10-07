<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    canvasVariablesStore,
    useVariable,
  } from "@rilldata/web-common/features/canvas/variables-store";
  import type { SelectProperties } from "@rilldata/web-common/features/templates/types";
  import type {
    V1ComponentSpecRendererProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";

  const MAX_OPTIONS = 250;
  const canvasName = getContext("rill::canvas:name") as string;

  export let componentName: string;
  export let data: any[] | undefined;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let output: V1ComponentVariable | undefined;

  $: outputVariableName = output?.name || "";
  $: outputVariableValue = useVariable(canvasName, outputVariableName);
  $: selectProperties = rendererProperties as SelectProperties;

  $: value = (value || $outputVariableValue || output?.defaultValue) as string;

  $: selectOptions = (data || [])
    .map((v) => ({
      value: String(v[selectProperties.valueField]),
      label: String(
        v[selectProperties?.labelField || selectProperties.valueField],
      ),
    }))
    .slice(0, MAX_OPTIONS);
</script>

<div>
  <Select
    on:change={(e) =>
      canvasVariablesStore.updateVariable(
        canvasName,
        outputVariableName,
        e.detail,
      )}
    bind:value
    id={componentName}
    tooltip={selectProperties.tooltip || ""}
    label={selectProperties.label || ""}
    options={selectOptions}
    placeholder={selectProperties.placeholder || ""}
  />
</div>
