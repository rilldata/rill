<script lang="ts">
  import type { CreateConnectorStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { connectorFormCache } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { onMount } from "svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import type { AddDataStateManager } from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";
  import { getEnvFileStore } from "@rilldata/web-common/features/env-management/env-file-store.ts";
  import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";

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

  const envStore = getEnvFileStore();

  let connectorName: string | null = null;
  let cachedEnvEditSession: EnvEditSession | null = null;
  let cachedFormValues: Record<string, unknown> | null = null;
  onMount(async () => {
    const { name, formValues, envEditSession } = connectorFormCache.getOrCreate(
      step.schema,
      step.connectorId,
      envStore,
    );
    connectorName = name;
    cachedFormValues = formValues;
    cachedEnvEditSession = envEditSession;
  });
</script>

{#if connectorName != null && cachedEnvEditSession != null && cachedFormValues != null}
  <ConnectorForm
    {stateManager}
    {step}
    {onSubmit}
    {onBack}
    {onClose}
    {connectorName}
    {cachedFormValues}
    {cachedEnvEditSession}
  />
{/if}
