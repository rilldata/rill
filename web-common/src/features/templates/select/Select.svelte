<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    dashboardVariablesStore,
    useVariable,
    useVariableInputParams,
  } from "@rilldata/web-common/features/custom-dashboards/variables-store";
  import { SelectProperties } from "@rilldata/web-common/features/templates/types";

  import {
    createQueryServiceResolveComponent,
    V1ComponentSpecRendererProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  const MAX_OPTIONS = 250;
  $: dashboardName = getContext("rill::custom-dashboard:name") as string;

  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let input: V1ComponentVariable[] | undefined;
  export let output: V1ComponentVariable | undefined;

  $: outputVariableName = output?.name || "";
  $: outputVariableValue = useVariable(dashboardName, outputVariableName);
  $: selectProperties = rendererProperties as SelectProperties;
  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: value = (value || $outputVariableValue || output?.defaultValue) as string;

  $: componentDataQuery = createQueryServiceResolveComponent(
    $runtime.instanceId,
    componentName,
    { args: $inputVariableParams },
  );

  $: selectOptions = ($componentDataQuery?.data?.data || [])
    .map((v) => ({
      value: String(v[selectProperties.valueField]),
      label: String(
        v[selectProperties?.labelField || selectProperties.valueField],
      ),
    }))
    .slice(0, MAX_OPTIONS);
</script>

<div class="m-1 p-1">
  <Select
    on:change={(e) =>
      dashboardVariablesStore.updateVariable(
        dashboardName,
        outputVariableName,
        e.detail,
      )}
    bind:value
    detach
    id={componentName}
    tooltip={selectProperties.tooltip || ""}
    label={selectProperties.label || ""}
    options={selectOptions}
    placeholder={selectProperties.placeholder || ""}
  />
</div>
