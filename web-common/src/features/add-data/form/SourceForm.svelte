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
    type AddDataConfig,
    type ImportAddDataStepConfig,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getImportStepsForSource } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import ConnectorHeader from "@rilldata/web-common/features/add-data/ConnectorHeader.svelte";
  import LocalSourceUpload from "@rilldata/web-common/features/sources/modal/LocalSourceUpload.svelte";
  import { getLabelsForSource } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import { uploadFile } from "@rilldata/web-common/features/sources/modal/file-upload.ts";
  import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";

  export let config: AddDataConfig;
  export let connectorDriver: V1ConnectorDriver;
  export let schemaName: string;
  export let connectorName: string;
  export let onSubmit: (importConfig: ImportAddDataStepConfig) => void;
  export let onBack: () => void;
  export let onClose: () => void;

  const runtimeClient = useRuntimeClient();

  const importSteps = getImportStepsForSource(config);

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

  $: sourceFormLabels = getLabelsForSource(importSteps);

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

    if (formValues.file) {
      // TODO: support multiple files upload
      const firstFile = formValues.file[0];
      const filePath = await uploadFile(runtimeClient, firstFile);
      if (filePath) {
        formValues.path = filePath;
        const [, fileName] = splitFolderFileNameAndExtension(filePath);
        formValues.name = getName(
          fileName,
          fileArtifacts.getNamesForKind(ResourceKind.Model),
        );
      }
    }
    const yaml = getSourceYamlPreview({
      connector: connectorDriver,
      formValues,
      schema,
      existingEnvBlob,
    });

    const importConfig = {
      importSteps,
      source: formValues.name,
      sourceSchema: "",
      sourceDatabase: "",
      connector: rewrittenConnector.name!,
      yaml,
      envBlob: newBlob,
    } satisfies ImportAddDataStepConfig;

    onSubmit(importConfig);
  }
</script>

<ConnectorHeader {connectorDriver} />

<AddDataFormStructure
  {connectorDriver}
  {schema}
  {superFormsParams}
  labels={sourceFormLabels}
  {yamlPreview}
  step="source"
  {onBack}
/>
