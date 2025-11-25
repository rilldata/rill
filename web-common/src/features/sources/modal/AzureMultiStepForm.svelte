<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";
  import { AZURE_AUTH_OPTIONS, type AzureAuthMethod } from "./constants";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;

  let azureAuthMethod: AzureAuthMethod = "connection_string";

  const filteredParamsProperties = properties;

  function clearFields(keys: string[]) {
    paramsForm.update(
      ($form) => {
        for (const key of keys) {
          $form[key] = "";
        }
        return $form;
      },
      { taint: false },
    );
  }

  $: if (azureAuthMethod === "connection_string") {
    clearFields([
      "azure_storage_account",
      "azure_storage_key",
      "azure_storage_sas_token",
    ]);
  } else if (azureAuthMethod === "storage_key") {
    clearFields(["azure_storage_connection_string", "azure_storage_sas_token"]);
  } else if (azureAuthMethod === "sas_token") {
    clearFields(["azure_storage_connection_string", "azure_storage_key"]);
  } else if (azureAuthMethod === "public") {
    clearFields([
      "azure_storage_connection_string",
      "azure_storage_account",
      "azure_storage_key",
      "azure_storage_sas_token",
    ]);
  }
</script>

<div>
  <div class="py-1.5 first:pt-0 last:pb-0">
    <div class="text-sm font-medium mb-4">Authentication</div>
    <Radio
      bind:value={azureAuthMethod}
      options={AZURE_AUTH_OPTIONS}
      name="azure-auth-method"
    >
      <svelte:fragment slot="custom-content" let:option>
        {#if option.value === "connection_string"}
          <Input
            id="azure_storage_connection_string"
            label="Connection String"
            placeholder="Enter your Azure storage connection string"
            optional={false}
            secret={true}
            hint="Paste an Azure Storage connection string"
            errors={normalizeErrors(
              paramsErrors.azure_storage_connection_string,
            )}
            bind:value={$paramsForm.azure_storage_connection_string}
            alwaysShowError
          />
        {:else if option.value === "storage_key"}
          <div class="space-y-3">
            <Input
              id="azure_storage_account"
              label="Storage Account Name"
              placeholder="Enter your Azure storage account name"
              optional={false}
              secret={false}
              hint="Exact name of the Azure storage account"
              errors={normalizeErrors(paramsErrors.azure_storage_account)}
              bind:value={$paramsForm.azure_storage_account}
              alwaysShowError
            />
            <Input
              id="azure_storage_key"
              label="Storage Account Key"
              placeholder="Enter your storage account key"
              optional={false}
              secret={true}
              hint="Primary or secondary storage account key"
              errors={normalizeErrors(paramsErrors.azure_storage_key)}
              bind:value={$paramsForm.azure_storage_key}
              alwaysShowError
            />
          </div>
        {:else if option.value === "sas_token"}
          <div class="space-y-3">
            <Input
              id="azure_storage_account"
              label="Storage Account Name"
              placeholder="Enter your Azure storage account name"
              optional={false}
              secret={false}
              hint="Account hosting the blob container"
              errors={normalizeErrors(paramsErrors.azure_storage_account)}
              bind:value={$paramsForm.azure_storage_account}
              alwaysShowError
            />
            <Input
              id="azure_storage_sas_token"
              label="SAS Token"
              placeholder="Enter your SAS token"
              optional={false}
              secret={true}
              hint="Shared Access Signature token with blob permissions"
              errors={normalizeErrors(paramsErrors.azure_storage_sas_token)}
              bind:value={$paramsForm.azure_storage_sas_token}
              alwaysShowError
            />
          </div>
        {/if}
      </svelte:fragment>
    </Radio>
  </div>

  {#each filteredParamsProperties as property (property.key)}
    {@const propertyKey = property.key ?? ""}
    {#if propertyKey !== "path" && propertyKey !== "azure_storage_connection_string" && propertyKey !== "azure_storage_account" && propertyKey !== "azure_storage_key" && propertyKey !== "azure_storage_sas_token"}
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
