<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import { normalizeErrors } from "./utils";
  import type { JSONSchemaField, MultiStepFormSchema } from "./types";
  import { isVisibleForValues } from "./multi-step-auth-configs";

  export let schema: MultiStepFormSchema | null = null;
  export let step: string | undefined = undefined;
  export let form: any;
  export let errors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  const radioDisplay = "radio";

  $: stepFilter = step;
  $: dependentMap = schema
    ? buildDependentMap(schema, stepFilter)
    : new Map<string, Array<[string, JSONSchemaField]>>();
  $: dependentKeys = new Set(
    Array.from(dependentMap.values()).flatMap((entries) =>
      entries.map(([key]) => key),
    ),
  );
  $: visibleEntries = schema
    ? computeVisibleEntries(schema, stepFilter, $form)
    : [];
  $: requiredFields = schema
    ? computeRequiredFields(schema, $form, stepFilter)
    : new Set<string>();
  $: renderOrder = schema
    ? computeRenderOrder(visibleEntries, dependentMap, dependentKeys)
    : [];

  // Seed defaults once when schema-provided defaults exist.
  $: if (schema) {
    form.update(
      ($form) => {
        const properties = schema.properties ?? {};
        for (const [key, prop] of Object.entries(properties)) {
          if (!matchesStep(prop, stepFilter)) continue;
          const current = $form[key];
          if (
            (current === undefined || current === null) &&
            prop.default !== undefined
          ) {
            $form[key] = prop.default;
          }
        }
        return $form;
      },
      { taint: false },
    );
  }

  // Clear hidden fields for the active step to avoid stale submissions.
  $: if (schema) {
    form.update(
      ($form) => {
        const properties = schema.properties ?? {};
        for (const [key, prop] of Object.entries(properties)) {
          if (!matchesStep(prop, stepFilter)) continue;
          const visible = isVisibleForValues(schema, key, $form);
          if (!visible && key in $form && $form[key] !== "") {
            $form[key] = "";
          }
        }
        return $form;
      },
      { taint: false },
    );
  }

  function matchesStep(prop: JSONSchemaField | undefined, stepValue?: string) {
    if (!stepValue) return true;
    const propStep = prop?.["x-step"];
    if (!propStep) return true;
    return propStep === stepValue;
  }

  function isRadioEnum(prop: JSONSchemaField) {
    return Boolean(prop.enum && prop["x-display"] === radioDisplay);
  }

  function computeVisibleEntries(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
    values: Record<string, unknown>,
  ) {
    const properties = currentSchema.properties ?? {};
    return Object.entries(properties).filter(([key, prop]) => {
      if (!matchesStep(prop, currentStep)) return false;
      return isVisibleForValues(currentSchema, key, values);
    });
  }

  function matchesCondition(
    condition:
      | Record<string, { const?: string | number | boolean }>
      | undefined,
    values: Record<string, unknown>,
  ) {
    if (!condition || !Object.keys(condition).length) return false;
    return Object.entries(condition).every(([depKey, def]) => {
      if (def.const === undefined || def.const === null) return false;
      return String(values?.[depKey]) === String(def.const);
    });
  }

  function computeRequiredFields(
    currentSchema: MultiStepFormSchema,
    values: Record<string, unknown>,
    currentStep: string | undefined,
  ) {
    const properties = currentSchema.properties ?? {};
    const required = new Set<string>();
    (currentSchema.required ?? []).forEach((key) => {
      if (matchesStep(properties[key], currentStep)) required.add(key);
    });

    for (const conditional of currentSchema.allOf ?? []) {
      const condition = conditional.if?.properties;
      const matches = matchesCondition(condition, values);
      const branch = matches ? conditional.then : conditional.else;
      branch?.required?.forEach((key) => {
        if (matchesStep(properties[key], currentStep)) required.add(key);
      });
    }
    return required;
  }

  function computeRenderOrder(
    entries: Array<[string, JSONSchemaField]>,
    dependents: Map<string, Array<[string, JSONSchemaField]>>,
    dependentKeySet: Set<string>,
  ) {
    const result: Array<[string, JSONSchemaField]> = [];
    const rendered = new Set<string>();

    for (const [key, prop] of entries) {
      if (rendered.has(key)) continue;

      if (isRadioEnum(prop)) {
        rendered.add(key);
        dependents.get(key)?.forEach(([childKey]) => rendered.add(childKey));
        result.push([key, prop]);
        continue;
      }

      if (dependentKeySet.has(key)) continue;

      rendered.add(key);
      result.push([key, prop]);
    }

    return result;
  }

  function buildDependentMap(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
  ) {
    const properties = currentSchema.properties ?? {};
    const map = new Map<string, Array<[string, JSONSchemaField]>>();

    for (const [key, prop] of Object.entries(properties)) {
      const visibleIf = prop["x-visible-if"];
      if (!visibleIf) continue;

      for (const controllerKey of Object.keys(visibleIf)) {
        const controller = properties[controllerKey];
        if (!controller) continue;
        if (!matchesStep(controller, currentStep)) continue;
        if (!matchesStep(prop, currentStep)) continue;

        const entries = map.get(controllerKey) ?? [];
        entries.push([key, prop]);
        map.set(controllerKey, entries);
      }
    }

    return map;
  }

  function getDependentFieldsForOption(
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    if (!schema) return [];
    const dependents = dependentMap.get(controllerKey) ?? [];
    const values = { ...$form, [controllerKey]: optionValue };
    return dependents.filter(([key]) =>
      isVisibleForValues(schema, key, values),
    );
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

  function isRequired(key: string) {
    return requiredFields.has(key);
  }
</script>

{#if schema}
  {#each renderOrder as [key, prop]}
    {#if isRadioEnum(prop)}
      <div class="py-1.5 first:pt-0 last:pb-0">
        {#if prop.title}
          <div class="text-sm font-medium mb-3">{prop.title}</div>
        {/if}
        <Radio
          bind:value={$form[key]}
          options={radioOptions(prop)}
          name={`${key}-radio`}
        >
          <svelte:fragment slot="custom-content" let:option>
            {#if dependentMap.get(key)?.length}
              {#each getDependentFieldsForOption(key, option.value) as [childKey, childProp]}
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
                  {:else if isRadioEnum(childProp)}
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
            {/if}
          </svelte:fragment>
        </Radio>
      </div>
    {:else}
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
        {:else if isRadioEnum(prop)}
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
    {/if}
  {/each}
{/if}
