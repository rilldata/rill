<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../runtime-client/v2";

  let {
    id = "connector-selector",
    value = "",
    onChange = () => {},
  }: {
    id?: string;
    value?: string;
    onChange?: (connector: string) => void;
  } = $props();

  const runtimeClient = useRuntimeClient();

  // Get all connectors that support SQL queries
  let connectorsQuery = $derived(
    createRuntimeServiceAnalyzeConnectors(
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
    ),
  );

  let options = $derived(
    ($connectorsQuery.data ?? []).map((c) => ({
      value: c.name as string,
      label: c.name as string,
    })),
  );
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
