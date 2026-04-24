<script lang="ts">
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import InlineConnectorForm from "./InlineConnectorForm.svelte";
  import type { InlineConnectorDriver } from "./writeInlineConnector";

  /** Maps option values to required driver names (from x-required-driver). */
  export let requiredDrivers: Record<string, string> = {};

  /** The currently selected option value. */
  export let currentValue: string | undefined;

  const SUPPORTED_DRIVERS: readonly InlineConnectorDriver[] = [
    "s3",
    "gcs",
    "azure",
  ];

  const client = useRuntimeClient();

  $: requiredDriver = currentValue
    ? requiredDrivers[currentValue]
    : undefined;

  $: isSupported =
    !!requiredDriver &&
    (SUPPORTED_DRIVERS as readonly string[]).includes(requiredDriver);

  $: analyzeQuery = createRuntimeServiceAnalyzeConnectors(
    client,
    {},
    {
      query: {
        enabled: isSupported,
        select: (data) => {
          const existing = new Set(
            (data.connectors ?? [])
              .map((c) => c.driver?.name ?? "")
              .filter(Boolean),
          );
          return existing;
        },
      },
    },
  );

  $: driverMissing =
    isSupported &&
    !!requiredDriver &&
    !!$analyzeQuery.data &&
    !$analyzeQuery.data.has(requiredDriver);
</script>

{#if driverMissing && requiredDriver}
  <div class="mt-2">
    <InlineConnectorForm driver={requiredDriver as InlineConnectorDriver} />
  </div>
{/if}
