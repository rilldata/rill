<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import { normalizeErrors } from "./utils";
  import type { JSONSchemaField, MultiStepFormSchema } from "./types";
  import {
    computeRequiredFields,
    isVisibleForValues,
    keysDependingOn,
    matchesStep,
    visibleFieldsForValues,
  } from "./schema-field-utils";

  export let schema: MultiStepFormSchema | null = null;
  export let step: string | null = null;
  export let form: any;
  export let errors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  $: properties = schema?.properties ?? {};

  // Apply defaults from the schema into the form when missing.
  $: if (schema && form) {
    const defaults = schema.properties ?? {};
    form.update(
      ($form) => {
        let mutated = false;
        const next = { ...$form };
        for (const [key, prop] of Object.entries(defaults)) {
          if (next[key] === undefined && prop.default !== undefined) {
            next[key] = prop.default;
            mutated = true;
          }
        }
        return mutated ? next : $form;
      },
      { taint: false },
    );
  }

  // Clear fields that are not visible for the current step to avoid
  // sending stale values for hidden inputs.
  $: if (schema && form) {
    form.update(
      ($form) => {
        let mutated = false;
        const next = { ...$form };
        for (const [key, prop] of Object.entries(properties)) {
          if (!matchesStep(prop, step)) continue;
          const visible = isVisibleForValues(schema, key, next);
          if (!visible && Object.prototype.hasOwnProperty.call(next, key)) {
            next[key] = "";
            mutated = true;
          }
        }
        return mutated ? next : $form;
      },
      { taint: false },
    );
  }

  $: requiredFields = schema
    ? computeRequiredFields(schema, { ...$form }, step)
    : new Set<string>();

  function isRequired(key: string) {
    return requiredFields.has(key);
  }

  function visibleFields(values: Record<string, unknown> = { ...$form }) {
    if (!schema) return [];
    return visibleFieldsForValues(schema, values, step);
  }

  function isRadioField(prop: JSONSchemaField) {
    return Boolean(prop.enum && prop["x-display"] === "radio");
  }

  function radioOptions(prop: JSONSchemaField) {
    return (
      prop.enum?.map((value, idx) => ({
        value: String(value),
        label: prop["x-enum-labels"]?.[idx] ?? String(value),
        description: prop["x-enum-descriptions"]?.[idx],
      })) ?? []
    );
  }

  $: visibleEntries = visibleFields();
  $: radioEntries = visibleEntries.filter(([, prop]) => isRadioField(prop));
  $: radioDependentKeys = schema
    ? keysDependingOn(
        schema,
        radioEntries.map(([key]) => key),
        step,
      )
    : new Set<string>();
  $: nonRadioEntries = visibleEntries.filter(
    ([key, prop]) => !isRadioField(prop) && !radioDependentKeys.has(key),
  );

  function visibleFieldsForRadioOption(
    fieldKey: string,
    optionValue: string | number | boolean,
  ) {
    if (!schema) return [];
    const values = { ...$form, [fieldKey]: optionValue };
    return visibleFieldsForValues(schema, values, step).filter(
      ([key, prop]) =>
        key !== fieldKey &&
        (radioDependentKeys.has(key) || isRadioField(prop)) &&
        matchesStep(prop, step),
    );
  }
</script>

{#if schema}
  {#each radioEntries as [key, prop]}
    <div class="py-1.5 first:pt-0 last:pb-0">
      <div class="text-sm font-medium mb-4">{prop.title ?? key}</div>
      <Radio
        bind:value={$form[key]}
        options={radioOptions(prop)}
        name={`${key}-radio`}
      >
        <svelte:fragment slot="custom-content" let:option>
          {#each visibleFieldsForRadioOption(key, option.value) as [childKey, childProp]}
            <div class="py-1.5 first:pt-0 last:pb-0">
              {#if childProp["x-display"] === "file" || childProp.format === "file"}
                <CredentialsInput
                  id={childKey}
                  hint={childProp.description ?? childProp["x-hint"]}
                  optional={!isRequired(childKey)}
                  bind:value={$form[childKey]}
                  uploadFile={handleFileUpload}
                  accept={childProp["x-accept"]}
                />
              {:else if childProp.type === "boolean"}
                <Checkbox
                  id={childKey}
                  bind:checked={$form[childKey]}
                  label={childProp.title ?? childKey}
                  hint={childProp.description ?? childProp["x-hint"]}
                  optional={!isRequired(childKey)}
                />
              {:else if isRadioField(childProp)}
                <Radio
                  bind:value={$form[childKey]}
                  options={radioOptions(childProp)}
                  name={`${childKey}-radio`}
                />
              {:else}
                <Input
                  id={childKey}
                  label={childProp.title ?? childKey}
                  placeholder={childProp["x-placeholder"]}
                  optional={!isRequired(childKey)}
                  secret={childProp["x-secret"]}
                  hint={childProp.description ?? childProp["x-hint"]}
                  errors={normalizeErrors(errors?.[childKey])}
                  bind:value={$form[childKey]}
                  onInput={(_, e) => onStringInputChange(e)}
                  alwaysShowError
                />
              {/if}
            </div>
          {/each}
        </svelte:fragment>
      </Radio>
    </div>
  {/each}

  {#each nonRadioEntries as [key, prop]}
    <div class="py-1.5 first:pt-0 last:pb-0">
      {#if prop["x-display"] === "file" || prop.format === "file"}
        <CredentialsInput
          id={key}
          hint={prop.description ?? prop["x-hint"]}
          optional={!isRequired(key)}
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
          optional={!isRequired(key)}
        />
      {:else if isRadioField(prop)}
        <Radio
          bind:value={$form[key]}
          options={radioOptions(prop)}
          name={`${key}-radio`}
        />
      {:else}
        <Input
          id={key}
          label={prop.title ?? key}
          placeholder={prop["x-placeholder"]}
          optional={!isRequired(key)}
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
