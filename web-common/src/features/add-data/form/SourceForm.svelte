<script lang="ts">
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import {
    runtimeServiceGetFile,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { onMount } from "svelte";
  import { getSourceYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { prepareSourceFormData } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import { getSchemaSecretKeys } from "@rilldata/web-common/features/templates/schema-utils.ts";
  import { updateDotEnvWithSecrets } from "@rilldata/web-common/features/connectors/code-utils.ts";
  import {
    type ImportAddDataStepConfig,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let connectorDriver: V1ConnectorDriver;
  export let schemaName: string;
  export let connectorName: string;
  export let onSubmit: (importConfig: ImportAddDataStepConfig) => void;
  export let onBack: () => void;

  const runtimeClient = useRuntimeClient();

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
    schemaName,
    formType: "source",
    onUpdate: async ({ form }) => {
      if (!form.valid) return;
      return submitImportConfig(form.data);
    },
  });

  $: ({ form } = superFormsParams);

  $: schema = getConnectorSchema(schemaName);
  $: yamlPreview = getSourceYamlPreview({
    connector: connectorDriver,
    formValues: $form,
    schema,
    existingEnvBlob,
  });

  async function submitImportConfig(formValues: any) {
    const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
      connectorDriver,
      formValues,
      { connectorInstanceName: connectorName },
    );
    const schema = getConnectorSchema(rewrittenConnector.name ?? "");
    const schemaSecretKeys = schema
      ? getSchemaSecretKeys(schema, { step: "source" })
      : [];

    // Create or update the `.env` file
    const { newBlob } = await updateDotEnvWithSecrets(
      runtimeClient,
      queryClient,
      rewrittenConnector,
      rewrittenFormValues,
      {
        secretKeys: schemaSecretKeys,
      },
    );

    const importConfig = {
      importSteps: [
        ImportDataStep.CreateModel,
        ImportDataStep.CreateMetricsView,
        ImportDataStep.CreateExplore,
      ],
      source: formValues.name,
      sourceSchema: "",
      sourceDatabase: "",
      connector: rewrittenConnector.name!,
      yaml: yamlPreview,
      envBlob: newBlob,
    } satisfies ImportAddDataStepConfig;

    onSubmit(importConfig);
  }
</script>

<AddDataFormStructure
  {connectorDriver}
  {schema}
  {superFormsParams}
  {yamlPreview}
  step="source"
  {onBack}
/>
