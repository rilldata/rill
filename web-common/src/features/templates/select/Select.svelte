<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { SelectProperties } from "@rilldata/web-common/features/templates/types";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  import {
    V1ComponentSpecRendererProperties,
    V1ComponentSpecResolverProperties,
  } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const MAX_OPTIONS = 250;

  export let componentName: string;
  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let resolverProperties: V1ComponentSpecResolverProperties | undefined;

  let value = "a";

  $: selectProperties = rendererProperties as SelectProperties;

  $: componentDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    componentName,
    {
      test: "test",
    },
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
    bind:value
    detach
    id={componentName}
    tooltip={selectProperties.tooltip || ""}
    label={selectProperties.label || ""}
    options={selectOptions}
    placeholder={selectProperties.placeholder || ""}
  />
</div>
