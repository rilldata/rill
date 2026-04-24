<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { fileArtifacts } from "../../entity-management/file-artifacts";
  import { getName } from "../../entity-management/name-utils";
  import { ResourceKind } from "../../entity-management/resource-selectors";
  import {
    writeInlineConnector,
    type InlineConnectorDriver,
  } from "./writeInlineConnector";

  export let driver: InlineConnectorDriver;

  const dispatch = createEventDispatcher<{
    success: { connectorName: string };
  }>();

  const client = useRuntimeClient();
  const queryClient = useQueryClient();

  const TITLES: Record<InlineConnectorDriver, string> = {
    s3: "Set up Amazon S3 credentials",
    gcs: "Set up Google Cloud Storage credentials",
    azure: "Set up Azure Blob Storage credentials",
  };

  const HINTS: Record<InlineConnectorDriver, string> = {
    s3: "Creates an S3 connector using access keys. For role-based auth, use the full S3 connector flow.",
    gcs: "Creates a GCS connector from a service account JSON. For HMAC auth, use the full GCS connector flow.",
    azure:
      "Creates an Azure connector from a connection string. For account key or SAS auth, use the full Azure connector flow.",
  };

  // Minimal happy-path fields per driver. If users need other auth methods,
  // they can still use the full connector flow.
  let awsAccessKeyId = "";
  let awsSecretAccessKey = "";
  let gcsCredentialsJson = "";
  let azureConnectionString = "";

  let submitting = false;
  let error: string | null = null;

  async function handleFileUpload(file: File): Promise<string> {
    const content = await file.text();
    // Return raw JSON; compileConnectorYAML will base64-encode secret values.
    gcsCredentialsJson = content;
    return content;
  }

  function getFormValues(): {
    values: Record<string, unknown>;
    valid: boolean;
  } {
    switch (driver) {
      case "s3":
        return {
          valid: !!awsAccessKeyId && !!awsSecretAccessKey,
          values: {
            auth_method: "access_keys",
            aws_access_key_id: awsAccessKeyId,
            aws_secret_access_key: awsSecretAccessKey,
          },
        };
      case "gcs":
        return {
          valid: !!gcsCredentialsJson,
          values: {
            auth_method: "credentials",
            google_application_credentials: gcsCredentialsJson,
          },
        };
      case "azure":
        return {
          valid: !!azureConnectionString,
          values: {
            auth_method: "connection_string",
            azure_storage_connection_string: azureConnectionString,
          },
        };
    }
  }

  async function handleSave() {
    const { values, valid } = getFormValues();
    if (!valid || submitting) return;

    submitting = true;
    error = null;

    try {
      const connectorName = getName(
        driver,
        fileArtifacts.getNamesForKind(ResourceKind.Connector),
      );

      await writeInlineConnector({
        client,
        queryClient,
        driver,
        values,
        connectorName,
      });

      dispatch("success", { connectorName });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      submitting = false;
    }
  }

  $: canSave = getFormValues().valid && !submitting;
</script>

<div class="inline-connector-form">
  <div class="header">
    <span class="title">{TITLES[driver]}</span>
    <span class="hint">{HINTS[driver]}</span>
  </div>

  <div class="fields">
    {#if driver === "s3"}
      <Input
        id="inline-s3-access-key-id"
        label="Access Key ID"
        placeholder="Enter AWS access key ID"
        secret
        bind:value={awsAccessKeyId}
        alwaysShowError
      />
      <Input
        id="inline-s3-secret-access-key"
        label="Secret Access Key"
        placeholder="Enter AWS secret access key"
        secret
        bind:value={awsSecretAccessKey}
        alwaysShowError
      />
    {:else if driver === "gcs"}
      <CredentialsInput
        id="inline-gcs-sa-json"
        label="Service account key"
        hint="Upload a JSON key file for a service account with GCS access."
        bind:value={gcsCredentialsJson}
        uploadFile={handleFileUpload}
        accept=".json"
      />
    {:else if driver === "azure"}
      <Input
        id="inline-azure-connection-string"
        label="Connection string"
        placeholder="Enter Azure storage connection string"
        secret
        bind:value={azureConnectionString}
        alwaysShowError
      />
    {/if}
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  <div class="actions">
    <Button
      type="primary"
      onClick={handleSave}
      disabled={!canSave}
      loading={submitting}
    >
      Save connector
    </Button>
  </div>
</div>

<style lang="postcss">
  .inline-connector-form {
    @apply flex flex-col gap-3;
    @apply p-3;
    @apply border border-gray-200 rounded;
    @apply bg-surface-secondary;
  }

  .header {
    @apply flex flex-col gap-0.5;
  }

  .title {
    @apply text-sm font-medium text-fg-primary;
  }

  .hint {
    @apply text-xs text-fg-muted;
  }

  .fields {
    @apply flex flex-col gap-2;
  }

  .actions {
    @apply flex justify-end;
  }

  .error {
    @apply text-xs text-red-600;
  }
</style>
