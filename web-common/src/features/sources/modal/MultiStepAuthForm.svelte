<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";

  export type AuthOption = {
    value: string;
    label: string;
    description: string;
    hint?: string;
  };

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;
  export let authOptions: AuthOption[] = [];
  export let defaultAuthMethod: string;
  export let clearFieldsByMethod: Record<string, string[]> = {};
  export let excludedKeys: string[] = [];

  // Keep auth method local to this component; default to provided value or first option.
  let authMethod: string = defaultAuthMethod || authOptions?.[0]?.value || "";

  // Reactive clearing of fields not relevant to the selected auth method.
  $: if (authMethod && clearFieldsByMethod[authMethod]?.length) {
    paramsForm.update(
      ($form) => {
        for (const key of clearFieldsByMethod[authMethod]) {
          if (key in $form) $form[key] = "";
        }
        return $form;
      },
      { taint: false },
    );
  }

  // Build a single exclusion set so we don't render auth fields twice.
  $: excluded = new Set([
    "path",
    ...excludedKeys,
    ...Object.values(clearFieldsByMethod).flat(),
  ]);
</script>

<!-- Auth method selector and dynamic auth fields -->
<div class="py-1.5 first:pt-0 last:pb-0">
  <div class="text-sm font-medium mb-4">Authentication method</div>
  <Radio bind:value={authMethod} options={authOptions} name="multi-auth-method">
    <svelte:fragment slot="custom-content" let:option>
      <slot
        name="auth-fields"
        {option}
        paramsFormStore={paramsForm}
        paramsErrors={paramsErrors}
        {handleFileUpload}
      />
    </svelte:fragment>
  </Radio>
</div>

<!-- Render other connector properties (excluding path and auth fields) -->
{#each properties as property (property.key)}
  {@const propertyKey = property.key ?? ""}
  {#if !excluded.has(propertyKey)}
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
