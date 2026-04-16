<script lang="ts">
  import type { CreateConnectorStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { connectorFormCache } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { onMount } from "svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import type { AddDataStateManager } from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";

  // Wrapper to initialize the ConnectorForm with cached data.
  // Has async logic to fetch the .env file. So to ensure we load the form on init, we use this wrapper.

  export let stateManager: AddDataStateManager;
  export let step: CreateConnectorStep;
  export let onSubmit: (
    connectorName: string,
    connectorFormValues: Record<string, unknown>,
  ) => void;
  export let onBack: () => void;
  export let onClose: () => void;

  let connectorName: string | null = null;
  let cachedEnvBlob: string | null = null;
  let cachedFormValues: Record<string, unknown> | null = null;
  onMount(async () => {
    const { name, formValues, existingEnvBlob } =
      await connectorFormCache.getOrCreate(step.schema, step.connectorId);
    connectorName = name;
    cachedEnvBlob = existingEnvBlob;
    cachedFormValues = formValues;
  });
</script>

{#if connectorName != null && cachedEnvBlob != null && cachedFormValues != null}
  <ConnectorForm
    {stateManager}
    {step}
    {onSubmit}
    {onBack}
    {onClose}
    {connectorName}
    {cachedFormValues}
    {cachedEnvBlob}
  />
{/if}
