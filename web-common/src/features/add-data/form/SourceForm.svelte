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
  import { submitAddSourceForm } from "@rilldata/web-common/features/sources/modal/submitAddDataForm.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  export let connectorDriver: V1ConnectorDriver;
  export let schemaName: string;
  export let connectorName: string;
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
    schemaName,
    formType: "source",
    onUpdate: async ({ form }) => {
      if (!form.valid) return;
      const newSourceName = await submitAddSourceForm(
        queryClient,
        connectorDriver,
        form.data,
        connectorName,
      );
      onSubmit(newSourceName);
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
</script>

<AddDataFormStructure {schema} {superFormsParams} {yamlPreview} {onBack} />
