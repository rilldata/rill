<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import { normalizeErrors } from "./utils";
  import type { MultiStepFormSchema } from "./types";
  import {
    findAuthMethodKey,
    getAuthOptionsFromSchema,
    getRequiredFieldsByAuthMethod,
    isVisibleForValues,
  } from "./multi-step-auth-configs";

  export let schema: MultiStepFormSchema | null = null;
  export let step: "connector" | "source" = "connector";
  export let form: any;
  export let errors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  // Bubble the selected auth method to the parent so it can adjust UI.
  export let authMethod: string = "";

  $: authInfo = schema ? getAuthOptionsFromSchema(schema) : null;
  $: authMethodKey = schema ? authInfo?.key || findAuthMethodKey(schema) : null;
  $: requiredByMethodConnector = schema
    ? getRequiredFieldsByAuthMethod(schema, { step: "connector" })
    : {};
  $: requiredByMethodSource = schema
    ? getRequiredFieldsByAuthMethod(schema, { step: "source" })
    : {};

  $: if (schema && authInfo && !authMethod) {
    authMethod = authInfo.defaultMethod || authInfo.options[0]?.value || "";
  }

  // Clear fields that are not visible for the active auth method to avoid
  // sending stale values across methods.
  $: if (schema && authMethod && step === "connector") {
    form.update(
      ($form) => {
        const properties = schema.properties ?? {};
        for (const key of Object.keys(properties)) {
          if (key === authMethodKey) continue;
          const prop = properties[key];
          const stepForField = prop["x-step"] ?? "connector";
          if (stepForField !== "connector") continue;
          const visible = isVisibleForValues(schema, key, {
            ...$form,
            [authMethodKey ?? "auth_method"]: authMethod,
          });
          if (!visible && key in $form) {
            $form[key] = "";
          }
        }
        return $form;
      },
      { taint: false },
    );
  }

  function visibleFieldsFor(
    method: string | undefined,
    currentStep: "connector" | "source",
  ) {
    if (!schema) return [];
    const properties = schema.properties ?? {};
    const values = { ...$form, [authMethodKey ?? "auth_method"]: method };
    return Object.entries(properties).filter(([key, prop]) => {
      if (authMethodKey && key === authMethodKey) return false;
      const stepForField = prop["x-step"] ?? "connector";
      if (stepForField !== currentStep) return false;
      return isVisibleForValues(schema, key, values);
    });
  }

  function isRequiredFor(method: string | undefined, key: string): boolean {
    if (!schema) return false;
    const requiredMap =
      step === "connector" ? requiredByMethodConnector : requiredByMethodSource;
    const requiredSet = requiredMap[method ?? ""] ?? [];
    return requiredSet.includes(key);
  }
</script>

{#if schema}
  {#if step === "connector" && authInfo}
    {#if authInfo.options.length > 1}
      <div class="py-1.5 first:pt-0 last:pb-0">
        <div class="text-sm font-medium mb-4">Authentication method</div>
        <Radio
          bind:value={authMethod}
          options={authInfo.options}
          name="multi-auth-method"
        >
          <svelte:fragment slot="custom-content" let:option>
            {#each visibleFieldsFor(option.value, "connector") as [key, prop]}
              <div class="py-1.5 first:pt-0 last:pb-0">
                {#if prop["x-display"] === "file" || prop.format === "file"}
                  <CredentialsInput
                    id={key}
                    hint={prop.description ?? prop["x-hint"]}
                    optional={!isRequiredFor(option.value, key)}
                    bind:value={$form[key]}
                    uploadFile={handleFileUpload}
                    accept={prop["x-accept"]}
                  />
                {:else if prop.type === "boolean"}
                  <Checkbox
                    id={key}
                    bind:checked={$form[key]}
                    label={prop.title ?? key}
                    hint={prop.description ?? prop["x-hint"]}
                    optional={!isRequiredFor(option.value, key)}
                  />
                {:else if prop.enum && prop["x-display"] === "radio"}
                  <Radio
                    bind:value={$form[key]}
                    options={prop.enum.map((value, idx) => ({
                      value: String(value),
                      label: prop["x-enum-labels"]?.[idx] ?? String(value),
                      description: prop["x-enum-descriptions"]?.[idx],
                    }))}
                    name={`${key}-radio`}
                  />
                {:else}
                  <Input
                    id={key}
                    label={prop.title ?? key}
                    placeholder={prop["x-placeholder"]}
                    optional={!isRequiredFor(option.value, key)}
                    secret={prop["x-secret"]}
                    hint={prop.description ?? prop["x-hint"]}
                    errors={normalizeErrors(errors?.[key])}
                    bind:value={$form[key]}
                    onInput={(_, e) => onStringInputChange(e)}
                    alwaysShowError
                  />
                {/if}
              </div>
            {/each}
          </svelte:fragment>
        </Radio>
      </div>
    {:else if authInfo.options[0]}
      {#each visibleFieldsFor(authMethod || authInfo.options[0].value, "connector") as [key, prop]}
        <div class="py-1.5 first:pt-0 last:pb-0">
          {#if prop["x-display"] === "file" || prop.format === "file"}
            <CredentialsInput
              id={key}
              hint={prop.description ?? prop["x-hint"]}
              optional={!isRequiredFor(authMethod, key)}
              bind:value={$form[key]}
              uploadFile={handleFileUpload}
              accept={prop["x-accept"]}
            />
          {:else if prop.type === "boolean"}
            <Checkbox
              id={key}
              bind:checked={$form[key]}
              label={prop.title ?? key}
              hint={prop.description ?? prop["x-hint"]}
              optional={!isRequiredFor(authMethod, key)}
            />
          {:else if prop.enum && prop["x-display"] === "radio"}
            <Radio
              bind:value={$form[key]}
              options={prop.enum.map((value, idx) => ({
                value: String(value),
                label: prop["x-enum-labels"]?.[idx] ?? String(value),
                description: prop["x-enum-descriptions"]?.[idx],
              }))}
              name={`${key}-radio`}
            />
          {:else}
            <Input
              id={key}
              label={prop.title ?? key}
              placeholder={prop["x-placeholder"]}
              optional={!isRequiredFor(authMethod, key)}
              secret={prop["x-secret"]}
              hint={prop.description ?? prop["x-hint"]}
              errors={normalizeErrors(errors?.[key])}
              bind:value={$form[key]}
              onInput={(_, e) => onStringInputChange(e)}
              alwaysShowError
            />
          {/if}
        </div>
      {/each}
    {/if}
  {:else}
    {#each visibleFieldsFor(authMethod, step) as [key, prop]}
      <div class="py-1.5 first:pt-0 last:pb-0">
        {#if prop["x-display"] === "file" || prop.format === "file"}
          <CredentialsInput
            id={key}
            hint={prop.description ?? prop["x-hint"]}
            optional={!isRequiredFor(authMethod, key)}
            bind:value={$form[key]}
            uploadFile={handleFileUpload}
            accept={prop["x-accept"]}
          />
        {:else if prop.type === "boolean"}
          <Checkbox
            id={key}
            bind:checked={$form[key]}
            label={prop.title ?? key}
            hint={prop.description ?? prop["x-hint"]}
            optional={!isRequiredFor(authMethod, key)}
          />
        {:else if prop.enum && prop["x-display"] === "radio"}
          <Radio
            bind:value={$form[key]}
            options={prop.enum.map((value, idx) => ({
              value: String(value),
              label: prop["x-enum-labels"]?.[idx] ?? String(value),
              description: prop["x-enum-descriptions"]?.[idx],
            }))}
            name={`${key}-radio`}
          />
        {:else}
          <Input
            id={key}
            label={prop.title ?? key}
            placeholder={prop["x-placeholder"]}
            optional={!isRequiredFor(authMethod, key)}
            secret={prop["x-secret"]}
            hint={prop.description ?? prop["x-hint"]}
            errors={normalizeErrors(errors?.[key])}
            bind:value={$form[key]}
            onInput={(_, e) => onStringInputChange(e)}
            alwaysShowError
          />
        {/if}
      </div>
    {/each}
  {/if}
{/if}
