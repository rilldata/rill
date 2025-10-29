<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;

  const filteredParamsProperties = properties;
</script>

<!-- Step 1: Connector configuration for S3 -->
<div>
  <div class="py-1.5 first:pt-0 last:pb-0">
    <div class="text-sm font-medium mb-4">Authentication</div>
    <div class="space-y-3">
      <Input
        id="aws_access_key_id"
        label="AWS access key ID"
        placeholder="Enter your AWS access key ID"
        optional={false}
        secret={true}
        hint="Access key ID for explicit credentials"
        errors={normalizeErrors(paramsErrors.aws_access_key_id)}
        bind:value={$paramsForm.aws_access_key_id}
        alwaysShowError
      />
      <Input
        id="aws_secret_access_key"
        label="AWS secret access key"
        placeholder="Enter your AWS secret access key"
        optional={false}
        secret={true}
        hint="Secret access key for explicit credentials"
        errors={normalizeErrors(paramsErrors.aws_secret_access_key)}
        bind:value={$paramsForm.aws_secret_access_key}
        alwaysShowError
      />
    </div>
  </div>

  <!-- Render other connector properties (excluding auth fields and path) -->
  {#each filteredParamsProperties as property (property.key)}
    {@const propertyKey = property.key ?? ""}
    {#if propertyKey !== "aws_access_key_id" && propertyKey !== "aws_secret_access_key" && propertyKey !== "aws_role_arn" && propertyKey !== "aws_role_session_name" && propertyKey !== "aws_external_id"}
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
