<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    createRuntimeServiceAnalyzeConnectors,
    createRuntimeServiceGetInstance,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";

  export let id: string = "connector-selector";
  export let value: string = "";
  export let onChange: (connector: string) => void = () => {};

  $: ({ instanceId } = $runtime);

  // Get the default OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instanceQuery.data?.instance?.olapConnector ?? "";

  // Set default connector when loaded
  $: if (olapConnector && !value) {
    value = olapConnector;
    onChange(olapConnector);
  }

  // Get all connectors that support SQL queries
  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      select: (data) => {
        if (!data?.connectors) return [];
        return data.connectors
          .filter(
            (c) =>
              c?.driver?.implementsOlap ||
              c?.driver?.implementsSqlStore ||
              c?.driver?.implementsWarehouse,
          )
          .sort((a, b) =>
            (a?.name as string).localeCompare(b?.name as string),
          );
      },
    },
  });

  $: options =
    ($connectorsQuery.data ?? []).map((c) => ({
      value: c.name as string,
      label: c.name as string,
    })) ?? [];
</script>

<Select
  {id}
  ariaLabel="Select connector"
  size="sm"
  {value}
  {options}
  optionsLoading={$connectorsQuery.isLoading}
  onChange={(newValue) => {
    value = newValue;
    onChange(newValue);
  }}
/>
