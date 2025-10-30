<script lang="ts">
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import {
    ConnectorDriverPropertyType,
    type ConnectorDriverProperty,
  } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";

  export let properties: Array<ConnectorDriverProperty> = [];
  export let form: any; // expect a store from parent
  export let errors: Record<string, any> | undefined;
  export let onStringInputChange: (event: Event) => void = () => {};
  export let uploadFile: (file: File) => Promise<string> = async () => "";
</script>

{#each properties as property (property.key)}
  {@const propertyKey = property.key ?? ""}
  {@const label =
    property.displayName + (property.required ? "" : " (optional)")}
  <div class="py-1.5 first:pt-0 last:pb-0">
    {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
      <Input
        id={propertyKey}
        label={property.displayName}
        placeholder={property.placeholder}
        optional={!property.required}
        secret={property.secret}
        hint={property.hint}
        errors={normalizeErrors(errors?.[propertyKey])}
        bind:value={$form[propertyKey]}
        onInput={(_, e) => onStringInputChange(e)}
        alwaysShowError
      />
    {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
      <Checkbox
        id={propertyKey}
        bind:checked={$form[propertyKey]}
        {label}
        hint={property.hint}
        optional={!property.required}
      />
    {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
      <InformationalField
        description={property.description}
        hint={property.hint}
        href={property.docsUrl}
      />
    {:else if property.type === ConnectorDriverPropertyType.TYPE_FILE}
      <CredentialsInput
        id={propertyKey}
        label={property.displayName}
        hint={property.hint}
        optional={!property.required}
        bind:value={$form[propertyKey]}
        {uploadFile}
        accept=".json"
      />
    {/if}
  </div>
{/each}
