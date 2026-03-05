<script lang="ts">
  import { createConnectorForm } from "@rilldata/web-common/features/sources/modal/FormValidation.ts";
  import {
    runtimeServiceGetFile,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { onMount } from "svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { getConnectorYamlPreview } from "./yaml-preview.ts";
  import AddDataFormStructure from "@rilldata/web-common/features/add-data/form/AddDataFormStructure.svelte";
  import { submitAddConnectorForm } from "@rilldata/web-common/features/sources/modal/submitAddDataForm.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  export let connectorDriver: V1ConnectorDriver;
  export let onSubmit: (name: string) => void;
  export let onBack: () => void;

  // Capture .env blob ONCE on mount for consistent conflict detection in YAML preview.
  // This prevents the preview from updating when Test and Connect writes to .env.
  // Use null to indicate "not yet loaded" vs "" for "loaded but empty"
  let existingEnvBlob: string | null = null;
  onMount(async () => {
    try {
      const envFile = await runtimeServiceGetFile($runtime.instanceId, {
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
      const connectorName = await submitAddConnectorForm(
        queryClient,
        connectorDriver,
        form.data,
        false,
        "",
        false,
      );
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

<AddDataFormStructure
  {schema}
  {superFormsParams}
  {yamlPreview}
  step="connector"
  {onBack}
/>
