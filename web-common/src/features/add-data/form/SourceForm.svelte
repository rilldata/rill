<script lang="ts">
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import {
    runtimeServiceGetFile,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { onMount } from "svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { getSourceYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { ImportTableRunner } from "@rilldata/web-common/features/add-data/import/ImportTableRunner.ts";
  import { prepareSourceFormData } from "@rilldata/web-common/features/sources/sourceUtils.ts";
  import { getSchemaSecretKeys } from "@rilldata/web-common/features/templates/schema-utils.ts";
  import { updateDotEnvWithSecrets } from "@rilldata/web-common/features/connectors/code-utils.ts";

  export let connectorDriver: V1ConnectorDriver;
  export let schemaName: string;
  export let connectorName: string;
  export let onSubmit: (runner: ImportTableRunner) => void;
  export let onBack: () => void;

  $: ({ instanceId } = $runtime);

  // Capture .env blob ONCE on mount for consistent conflict detection in YAML preview.
  // This prevents the preview from updating when Test and Connect writes to .env.
  // Use null to indicate "not yet loaded" vs "" for "loaded but empty"
  let existingEnvBlob: string | null = null;
  onMount(async () => {
    try {
      const envFile = await runtimeServiceGetFile(instanceId, {
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
      const runner = await getRunner(form.data);
      onSubmit(runner);
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

  async function getRunner(formValues: any) {
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
      queryClient,
      rewrittenConnector,
      rewrittenFormValues,
      {
        secretKeys: schemaSecretKeys,
      },
    );

    return new ImportTableRunner(
      instanceId,
      formValues.name,
      {
        connector: rewrittenConnector.name,
        schema: "",
        database: "",
        table: "",
      },
      yamlPreview,
      newBlob,
    );
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
