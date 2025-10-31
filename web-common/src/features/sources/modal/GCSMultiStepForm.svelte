<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";
  import { GCS_AUTH_OPTIONS, type GCSAuthMethod } from "./constants";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  let gcsAuthMethod: GCSAuthMethod = "credentials";

  const filteredParamsProperties = properties;

  // Clear irrelevant auth fields when switching methods so preview doesn't retain stale values
  $: if (gcsAuthMethod === "hmac") {
    // Switching to HMAC: remove service account JSON from form
    paramsForm.update(
      ($form) => {
        $form.google_application_credentials = "";
        return $form;
      },
      { taint: false },
    );
  } else if (gcsAuthMethod === "credentials") {
    // Switching to Credentials: clear HMAC fields
    paramsForm.update(
      ($form) => {
        $form.key_id = "";
        $form.secret = "";
        return $form;
      },
      { taint: false },
    );
  }
</script>

<!-- Step 1: Connector configuration -->
<div>
  <div class="py-1.5 first:pt-0 last:pb-0">
    <div class="text-sm font-medium mb-4">Authentication method</div>
    <Radio
      bind:value={gcsAuthMethod}
      options={GCS_AUTH_OPTIONS}
      name="gcs-auth-method"
    >
      <svelte:fragment slot="custom-content" let:option>
        {#if option.value === "credentials"}
          <CredentialsInput
            id="google_application_credentials"
            hint="Upload a JSON key file for a service account with GCS access."
            optional={false}
            bind:value={$paramsForm.google_application_credentials}
            uploadFile={handleFileUpload}
            accept=".json"
          />
        {:else if option.value === "hmac"}
          <div class="space-y-3">
            <Input
              id="key_id"
              label="Access Key ID"
              placeholder="Enter your HMAC access key ID"
              optional={false}
              secret={true}
              hint="HMAC access key ID for S3-compatible authentication"
              errors={normalizeErrors(paramsErrors.key_id)}
              bind:value={$paramsForm.key_id}
              alwaysShowError
            />
            <Input
              id="secret"
              label="Secret Access Key"
              placeholder="Enter your HMAC secret access key"
              optional={false}
              secret={true}
              hint="HMAC secret access key for S3-compatible authentication"
              errors={normalizeErrors(paramsErrors.secret)}
              bind:value={$paramsForm.secret}
              alwaysShowError
            />
          </div>
        {/if}
      </svelte:fragment>
    </Radio>
  </div>

  <!-- Render other connector properties (excluding path and auth fields) -->
  {#each filteredParamsProperties as property (property.key)}
    {@const propertyKey = property.key ?? ""}
    {#if propertyKey !== "path" && propertyKey !== "google_application_credentials" && propertyKey !== "key_id" && propertyKey !== "secret"}
      <div class="py-1.5 first:pt-0 last:pb-0">
        {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
          <Input
            id={propertyKey}
            label={property.displayName}
            placeholder={property.placeholder}
            optional={!property.required}
            secret={property.secret}
            hint={property.hint}
            errors={normalizeErrors(paramsErrors[propertyKey])}
            bind:value={$paramsForm[propertyKey]}
            onInput={(_, e) => onStringInputChange(e)}
            alwaysShowError
          />
        {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
          <Checkbox
            id={propertyKey}
            bind:checked={$paramsForm[propertyKey]}
            label={property.displayName}
            hint={property.hint}
            optional={!property.required}
          />
        {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
          <InformationalField
            description={property.description}
            hint={property.hint}
            href={property.docsUrl}
          />
        {/if}
      </div>
    {/if}
  {/each}
</div>
