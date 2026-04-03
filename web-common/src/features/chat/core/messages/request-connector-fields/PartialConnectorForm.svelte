<script lang="ts">
  import { onMount } from "svelte";
  import { getConnectorDriverForSchema } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { createConnector } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { setSubmitError } from "@rilldata/web-common/features/add-data/form/errors.ts";
  import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { getLabelsForConnector } from "@rilldata/web-common/features/add-data/form/form-labels.ts";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";

  export let schemaName: string;
  export let connectorPath: string | undefined;
  export let schema: MultiStepFormSchema;
  export let existingData: Record<string, any>;
  export let onSubmit: (newConnectorPath: string) => void;

  const runtimeClient = getRuntimeClient();

  const connectorName = connectorPath
    ? getNameFromFile(connectorPath)
    : getName(
        schemaName,
        fileArtifacts.getNamesForKind(ResourceKind.Connector),
      );

  $: connectorDriver = getConnectorDriverForSchema(schemaName);

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
    formType: "connector",
    onUpdate: async ({ form }) => {
      if (!form.valid || !connectorDriver) return;
      try {
        const newConnectorPath = await createConnector({
          runtimeClient,
          queryClient,
          connectorName,
          connectorDriver,
          formValues: {
            ...form.data,
            ...existingData,
          },
          validate: false,
          existingEnvBlob,
          connectorPath,
        });

        onSubmit(newConnectorPath);
      } catch (e) {
        setSubmitError(form, e);
      }
    },
    schemaOverride: schema,
  });

  $: ({ form } = superFormsParams);

  $: labelsForConnector = getLabelsForConnector(schema, $form);
</script>

<AddDataFormStructure
  {connectorDriver}
  {schema}
  {superFormsParams}
  labels={labelsForConnector}
  step={{
    type: AddDataStep.CreateConnector,
    schema: schemaName,
    connectorId: connectorName,
  }}
/>
