<script lang="ts">
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import {
    runtimeServiceGetFile,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { onMount } from "svelte";
  import { getConnectorYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ConnectorHeader from "@rilldata/web-common/features/add-data/ConnectorHeader.svelte";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createConnector } from "@rilldata/web-common/features/add-data/steps/connector.ts";

  export let connectorDriver: V1ConnectorDriver;
  export let onSubmit: (name: string) => void;
  export let onBack: () => void;

  const runtimeClient = useRuntimeClient();
  const connectorName = getName(
    connectorDriver.name!,
    fileArtifacts.getNamesForKind(ResourceKind.Connector),
  );

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
    schemaName: connectorDriver.name ?? "",
    formType: "connector",
    onUpdate: async ({ form }) => {
      if (!form.valid) return;
      await createConnector({
        runtimeClient,
        queryClient,
        connectorName,
        connectorDriver,
        formValues: form.data,
        saveAnyway: false,
      });
      onSubmit(connectorName);
    },
  });

  $: ({ form } = superFormsParams);

  $: schema = getConnectorSchema(connectorDriver.name ?? "");
  $: yamlPreview = getConnectorYamlPreview({
    connector: connectorDriver,
    formValues: $form,
    schema,
    existingEnvBlob,
  });
</script>

<ConnectorHeader {connectorDriver} />

<AddDataFormStructure
  {connectorDriver}
  {schema}
  {superFormsParams}
  {yamlPreview}
  step="connector"
  {onBack}
/>
