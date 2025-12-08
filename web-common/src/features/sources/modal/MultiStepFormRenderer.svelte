<script lang="ts">
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiStepAuthForm from "./MultiStepAuthForm.svelte";
  import { normalizeErrors } from "./utils";
  import { type MultiStepFormConfig } from "./types";

  export let config: MultiStepFormConfig | null = null;
  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;
</script>

{#if config}
  <MultiStepAuthForm
    {properties}
    {paramsForm}
    {paramsErrors}
    {onStringInputChange}
    {handleFileUpload}
    authOptions={config.authOptions}
    defaultAuthMethod={config.defaultAuthMethod ||
      config.authOptions?.[0]?.value}
    clearFieldsByMethod={config.clearFieldsByMethod}
    excludedKeys={config.excludedKeys}
  >
    <svelte:fragment
      slot="auth-fields"
      let:option
      let:paramsErrors
      let:handleFileUpload
    >
      {#if config.authFieldGroups?.[option.value]}
        <div class="space-y-3">
          {#each config.authFieldGroups[option.value] as field (field.id)}
            {#if field.type === "credentials"}
              <CredentialsInput
                id={field.id}
                hint={field.hint}
                optional={field.optional ?? false}
                bind:value={$paramsForm[field.id]}
                uploadFile={handleFileUpload}
                accept={field.accept}
              />
            {:else}
              <Input
                id={field.id}
                label={field.label}
                placeholder={field.placeholder}
                optional={field.optional ?? false}
                secret={field.secret}
                hint={field.hint}
                errors={normalizeErrors(paramsErrors[field.id])}
                bind:value={$paramsForm[field.id]}
                alwaysShowError
              />
            {/if}
          {/each}
        </div>
      {/if}
    </svelte:fragment>
  </MultiStepAuthForm>
{/if}
