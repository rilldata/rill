<script lang="ts">
  import { goto } from "$app/navigation";
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { getConnectorYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    connectorFormCache,
    createConnector,
    maybeDeleteConnector,
  } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { getLabelsForConnector } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import { setSubmitError } from "@rilldata/web-common/features/add-data/form/errors.ts";
  import type {
    AddDataConfig,
    CreateConnectorStep,
  } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { getConnectorDriverForSchema } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
  import type { AddDataStateManager } from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";

  export let config: AddDataConfig;
  export let stateManager: AddDataStateManager;
  export let step: CreateConnectorStep;
  export let onSubmit: (
    connectorName: string,
    connectorFormValues: Record<string, unknown>,
  ) => void;
  export let onBack: () => void;
  export let onClose: () => void;

  export let cachedFormValues: Record<string, unknown>;
  export let connectorName: string;
  export let cachedEnvBlob: string;

  const runtimeClient = useRuntimeClient();

  $: connectorDriver = getConnectorDriverForSchema(step.schema);

  const superFormsParams = createConnectorForm({
    schemaName: step.schema,
    formType: "connector",
    onUpdate: async ({ form }) => {
      if (!form.valid || !connectorDriver) return;
      try {
        await createConnector({
          runtimeClient,
          queryClient,
          connectorName,
          connectorDriver,
          formValues: form.data,
          validate: true,
          existingEnvBlob: cachedEnvBlob,
        });

        connectorFormCache.updateFormValues(step.connectorId, form.data);

        onSubmit(connectorName, form.data);
      } catch (e) {
        stateManager.fireErrorEvent(e.message);
        setSubmitError(form, e);
      }
    },
    additionalDefaults: cachedFormValues,
  });

  $: ({ form } = superFormsParams);

  $: schema = getConnectorSchema(step.schema);
  $: yamlPreview = connectorDriver
    ? getConnectorYamlPreview({
        connector: connectorDriver,
        formValues: $form,
        schema,
        existingEnvBlob: cachedEnvBlob,
      })
    : "";

  $: labelsForConnector = getLabelsForConnector(schema, $form);

  async function saveConnector() {
    if (!connectorDriver) return;
    const connectorPath = await createConnector({
      runtimeClient,
      queryClient,
      connectorName,
      connectorDriver,
      formValues: $form,
      validate: false,
      existingEnvBlob: cachedEnvBlob,
    });
    onClose();
    if (!config.skipNavigation)
      return goto(`/files${addLeadingSlash(connectorPath)}`);
  }

  async function cleanupAndBack() {
    await maybeDeleteConnector(runtimeClient, queryClient, connectorName);

    onBack();
  }
</script>

{#if connectorDriver}
  <AddDataFormStructure
    {connectorDriver}
    {schema}
    {superFormsParams}
    labels={labelsForConnector}
    {yamlPreview}
    {step}
    onSave={saveConnector}
    onBack={cleanupAndBack}
  />
{/if}
