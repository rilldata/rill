<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiStepAuthForm from "./MultiStepAuthForm.svelte";
  import { normalizeErrors } from "./utils";
  import { AZURE_AUTH_OPTIONS, type AzureAuthMethod } from "./constants";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  const filteredParamsProperties = properties;

  const AZURE_CLEAR_FIELDS: Record<AzureAuthMethod, string[]> = {
    account_key: ["azure_storage_connection_string", "azure_storage_sas_token"],
    sas_token: ["azure_storage_connection_string", "azure_storage_key"],
    connection_string: [
      "azure_storage_account",
      "azure_storage_key",
      "azure_storage_sas_token",
    ],
  };

  const AZURE_EXCLUDED_KEYS = [
    "azure_storage_account",
    "azure_storage_key",
    "azure_storage_sas_token",
    "azure_storage_connection_string",
  ];
</script>

<MultiStepAuthForm
  properties={filteredParamsProperties}
  {paramsForm}
  {paramsErrors}
  {onStringInputChange}
  {handleFileUpload}
  authOptions={AZURE_AUTH_OPTIONS}
  defaultAuthMethod="account_key"
  clearFieldsByMethod={AZURE_CLEAR_FIELDS}
  excludedKeys={AZURE_EXCLUDED_KEYS}
>
  <svelte:fragment slot="auth-fields" let:option let:paramsErrors>
    {#if option.value === "account_key"}
      <div class="space-y-3">
        <Input
          id="azure_storage_account"
          label="Storage account"
          placeholder="Enter Azure storage account"
          optional={false}
          secret={false}
          hint="The name of the Azure storage account"
          errors={normalizeErrors(paramsErrors.azure_storage_account)}
          bind:value={$paramsForm.azure_storage_account}
          alwaysShowError
        />
        <Input
          id="azure_storage_key"
          label="Access key"
          placeholder="Enter Azure storage access key"
          optional={false}
          secret={true}
          hint="Primary or secondary access key for the storage account"
          errors={normalizeErrors(paramsErrors.azure_storage_key)}
          bind:value={$paramsForm.azure_storage_key}
          alwaysShowError
        />
      </div>
    {:else if option.value === "sas_token"}
      <div class="space-y-3">
        <Input
          id="azure_storage_account"
          label="Storage account"
          placeholder="Enter Azure storage account"
          optional={false}
          secret={false}
          errors={normalizeErrors(paramsErrors.azure_storage_account)}
          bind:value={$paramsForm.azure_storage_account}
          alwaysShowError
        />
        <Input
          id="azure_storage_sas_token"
          label="SAS token"
          placeholder="Enter Azure SAS token"
          optional={false}
          secret={true}
          hint="Shared Access Signature token for the storage account"
          errors={normalizeErrors(paramsErrors.azure_storage_sas_token)}
          bind:value={$paramsForm.azure_storage_sas_token}
          alwaysShowError
        />
      </div>
    {:else if option.value === "connection_string"}
      <Input
        id="azure_storage_connection_string"
        label="Connection string"
        placeholder="Enter Azure storage connection string"
        optional={false}
        secret={true}
        hint="Full connection string for the storage account"
        errors={normalizeErrors(paramsErrors.azure_storage_connection_string)}
        bind:value={$paramsForm.azure_storage_connection_string}
        alwaysShowError
      />
    {/if}
  </svelte:fragment>
</MultiStepAuthForm>
