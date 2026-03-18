<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../runtime-client/v2";

  export let id: string = "connector-selector";
  export let value: string = "";
  export let onChange: (connector: string) => void = () => {};

  const runtimeClient = useRuntimeClient();

  // Get all connectors that support SQL queries
  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(
    runtimeClient,
    {},
    {
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
            .sort((a, b) => (a?.name ?? "").localeCompare(b?.name ?? ""));
        },
      },
    },
  );

  $: options = ($connectorsQuery.data ?? []).map((c) => ({
    value: c.name as string,
    label: c.name as string,
  }));
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
