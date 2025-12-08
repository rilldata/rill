<script lang="ts">
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { normalizeErrors } from "./utils";
  import MultiStepAuthForm from "./MultiStepAuthForm.svelte";
  import { GCS_AUTH_OPTIONS, type GCSAuthMethod } from "./constants";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;
  const filteredParamsProperties = properties;
  const GCS_CLEAR_FIELDS: Record<GCSAuthMethod, string[]> = {
    credentials: ["key_id", "secret"],
    hmac: ["google_application_credentials"],
  };
  const GCS_EXCLUDED_KEYS = [
    "google_application_credentials",
    "key_id",
    "secret",
  ];
</script>

<MultiStepAuthForm
  properties={filteredParamsProperties}
  {paramsForm}
  {paramsErrors}
  {onStringInputChange}
  {handleFileUpload}
  authOptions={GCS_AUTH_OPTIONS}
  defaultAuthMethod="credentials"
  clearFieldsByMethod={GCS_CLEAR_FIELDS}
  excludedKeys={GCS_EXCLUDED_KEYS}
>
  <svelte:fragment
    slot="auth-fields"
    let:option
    let:paramsErrors
    let:handleFileUpload
  >
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
</MultiStepAuthForm>
