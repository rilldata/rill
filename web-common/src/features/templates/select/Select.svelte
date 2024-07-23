<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    dashboardVariablesStore,
    useVariableInputParams,
  } from "@rilldata/web-common/features/custom-dashboards/variables-store";
  import { SelectProperties } from "@rilldata/web-common/features/templates/types";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext } from "svelte";

  const MAX_OPTIONS = 250;
  const dashboardName = getContext("rill::custom-dashboard:name") as string;

  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;
  export let input: V1ComponentVariable[] | undefined;
  export let output: V1ComponentVariable | undefined;

  let value = output?.defaultValue as string;

  $: outputVariableName = output?.name || "";
  $: selectProperties = rendererProperties as SelectProperties;
  $: inputVariableParams = useVariableInputParams(dashboardName, input);

  $: componentDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    componentName,
    $inputVariableParams,
    resolverProperties,
  );

  $: selectOptions = ($componentDataQuery?.data || [])
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
