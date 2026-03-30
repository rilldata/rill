<script lang="ts">
  import { goto } from "$app/navigation";
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { onMount } from "svelte";
  import { getConnectorYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createConnector,
    maybeDeleteConnector,
  } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { getLabelsForConnector } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import { setSubmitError } from "@rilldata/web-common/features/add-data/form/errors.ts";
  import type { CreateConnectorStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { getConnectorDriverForSchema } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";

  export let step: CreateConnectorStep;
  export let onSubmit: (
    connectorName: string,
    connectorFormValues: Record<string, unknown>,
  ) => void;
  export let onBack: () => void;
  export let onClose: () => void;

  const runtimeClient = useRuntimeClient();

  $: connectorDriver = getConnectorDriverForSchema(step.schema);

  // Capture .env blob ONCE on mount for consistent conflict detection in YAML preview.
  // This prevents the preview from updating when Test and Connect writes to .env.
  // Use null to indicate "not yet loaded" vs "" for "loaded but empty"
  let existingEnvBlob: string | null = null;
  onMount(async () => {
    try {
      const envFile = await runtimeServiceGetFile(runtimeClient, {
        path: ".env",
      });
      existingEnvBlob = envFile.blob ?? "";
    } catch {
      // .env doesn't exist yet
      existingEnvBlob = "";
    }
  });

  const superFormsParams = createConnectorForm({
    schemaName: step.schema,
    formType: "connector",
    onUpdate: async ({ form }) => {
      if (!form.valid || !connectorDriver) return;
      try {
        await createConnector({
          runtimeClient,
          queryClient,
          connectorName: step.assignedConnectorName,
          connectorDriver,
          formValues: form.data,
          validate: true,
          existingEnvBlob,
        });
        onSubmit(step.assignedConnectorName, form.data);
      } catch (e) {
        setSubmitError(form, e);
      }
    },
  });

  $: ({ form } = superFormsParams);

  $: schema = getConnectorSchema(step.schema);
  $: yamlPreview = connectorDriver
    ? getConnectorYamlPreview({
        connector: connectorDriver,
        formValues: $form,
        schema,
        existingEnvBlob,
      })
    : "";

  $: labelsForConnector = getLabelsForConnector(schema, $form);

  async function saveConnector() {
    if (!connectorDriver) return;
    const connectorPath = await createConnector({
      runtimeClient,
      queryClient,
      connectorName: step.assignedConnectorName,
      connectorDriver,
      formValues: $form,
      validate: false,
      existingEnvBlob,
    });
    onClose();
    return goto(`/files${addLeadingSlash(connectorPath)}`);
  }

  async function cleanupAndBack() {
    await maybeDeleteConnector(
      runtimeClient,
      queryClient,
      step.assignedConnectorName,
    );

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
    step="connector"
    onSave={saveConnector}
    onBack={cleanupAndBack}
  />
{/if}
