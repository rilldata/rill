<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    dashboardVariablesStore,
    useVariable,
  } from "@rilldata/web-common/features/canvas-dashboards/variables-store";
  import { SwitchProperties } from "@rilldata/web-common/features/templates/types";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { getContext } from "svelte";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let output: V1ComponentVariable | undefined;

  $: dashboardName = getContext("rill::canvas-dashboard:name") as string;
  $: outputVariableName = output?.name || "";
  $: outputVariableValue = useVariable(dashboardName, outputVariableName);
  $: switchProperties = rendererProperties as SwitchProperties;

  $: value = (value || $outputVariableValue || output?.defaultValue) as boolean;
</script>

<Tooltip
  distance={8}
  location="bottom"
  alignment="start"
  suppress={!switchProperties?.tooltip}
>
  <slot name="tooltip" />
  <div class="m-1 p-1 flex items-center h-full">
    <Switch
      checked={value}
      on:click={() => {
        value = !value;
        dashboardVariablesStore.updateVariable(
          dashboardName,
          outputVariableName,
          value,
        );
      }}
    >
      {switchProperties.label}
    </Switch>
  </div>
  <TooltipContent slot="tooltip-content">
    {switchProperties?.tooltip}
  </TooltipContent>
</Tooltip>
