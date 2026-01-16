<script lang="ts">
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import SchemaField from "./SchemaField.svelte";
  import type { JSONSchemaField, MultiStepFormSchema } from "./schemas/types";
  import { isStepMatch, isVisibleForValues } from "./schema-utils";

  export let schema: MultiStepFormSchema | null = null;
  export let step: string | undefined = undefined;
  export let form: any;
  export let errors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  const radioDisplay = "radio";
  const tabsDisplay = "tabs";

  $: stepFilter = step;
  $: groupedFields = schema
    ? buildGroupedFields(schema, stepFilter)
    : new Map<string, Record<string, string[]>>();
  $: tabGroupedFields = schema
    ? buildTabGroupedFields(schema, stepFilter)
    : new Map<string, Record<string, string[]>>();
  $: groupedChildKeys = new Set(
    [
      ...Array.from(groupedFields.values()).flatMap((group) =>
        Object.values(group).flat(),
      ),
      ...Array.from(tabGroupedFields.values()).flatMap((group) =>
        Object.values(group).flat(),
      ),
    ],
  );
  $: visibleEntries = schema
    ? computeVisibleEntries(schema, stepFilter, $form)
    : [];
  $: requiredFields = schema
    ? computeRequiredFields(schema, $form, stepFilter)
    : new Set<string>();
  $: renderOrder = schema
    ? computeRenderOrder(visibleEntries, groupedChildKeys)
    : [];

  // Seed defaults for initial render: use explicit defaults, and for radio enums
  // fall back to first option when no value is set.
  $: if (schema) {
    form.update(
      ($form) => {
        const properties = schema.properties ?? {};
        for (const [key, prop] of Object.entries(properties)) {
          if (!isStepMatch(schema, key, stepFilter)) continue;
          const current = $form[key];
          const isUnset =
            current === undefined || current === null || current === "";

          if (isUnset && prop.default !== undefined) {
            $form[key] = prop.default;
          } else if (isUnset && isRadioEnum(prop) && prop.enum?.length) {
            $form[key] = String(prop.enum[0]);
          } else if (isUnset && isTabsEnum(prop) && prop.enum?.length) {
            $form[key] = String(prop.enum[0]);
          }
        }
        return $form;
      },
      { taint: false },
    );
  }

  // Clear hidden fields for the active step to avoid stale submissions.
  // Depend on `$form` so this runs when the auth method (or other values) change.
  $: if (schema) {
    const currentValues = $form;
    const properties = schema.properties ?? {};

    const shouldClear = Object.entries(properties).some(([key]) => {
      if (!isStepMatch(schema, key, stepFilter)) return false;
      const visible = isVisibleForValues(schema, key, currentValues);
      return !visible && key in currentValues && currentValues[key] !== "";
    });

    if (shouldClear) {
      form.update(
        ($form) => {
          for (const key of Object.keys(properties)) {
            if (!isStepMatch(schema, key, stepFilter)) continue;
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
  }

  function isRadioEnum(prop: JSONSchemaField) {
    return Boolean(prop.enum && prop["x-display"] === radioDisplay);
  }

  function isTabsEnum(prop: JSONSchemaField) {
    return Boolean(prop.enum && prop["x-display"] === tabsDisplay);
  }

  function computeVisibleEntries(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
    values: Record<string, unknown>,
  ) {
    const properties = currentSchema.properties ?? {};
    return Object.entries(properties).filter(([key]) => {
      if (!isStepMatch(currentSchema, key, currentStep)) return false;
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
    const required = new Set<string>();
    (currentSchema.required ?? []).forEach((key) => {
      if (isStepMatch(currentSchema, key, currentStep)) required.add(key);
    });

    for (const conditional of currentSchema.allOf ?? []) {
      const condition = conditional.if?.properties;
      const matches = matchesCondition(condition, values);
      const branch = matches ? conditional.then : conditional.else;
      branch?.required?.forEach((key) => {
        if (isStepMatch(currentSchema, key, currentStep)) required.add(key);
      });
    }
    return required;
  }

  function computeRenderOrder(
    entries: Array<[string, JSONSchemaField]>,
    groupedChildKeySet: Set<string>,
  ) {
    const result: Array<[string, JSONSchemaField]> = [];
    for (const [key, prop] of entries) {
      if (groupedChildKeySet.has(key)) continue;
      result.push([key, prop]);
    }

    return result;
  }

  function buildGroupedFields(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
  ): Map<string, Record<string, string[]>> {
    const properties = currentSchema.properties ?? {};
    const map = new Map<string, Record<string, string[]>>();

    for (const [key, prop] of Object.entries(properties)) {
      const grouped = prop["x-grouped-fields"];
      if (!grouped) continue;
      if (!isStepMatch(currentSchema, key, currentStep)) continue;

      const filteredOptions: Record<string, string[]> = {};
      const groupedEntries = Object.entries(grouped) as Array<
        [string, string[]]
      >;
      for (const [optionValue, childKeys] of groupedEntries) {
        filteredOptions[optionValue] = childKeys.filter((childKey) => {
          const childProp = properties[childKey];
          if (!childProp) return false;
          return isStepMatch(currentSchema, childKey, currentStep);
        });
      }
      map.set(key, filteredOptions);
    }

    return map;
  }

  function buildTabGroupedFields(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
  ): Map<string, Record<string, string[]>> {
    const properties = currentSchema.properties ?? {};
    const map = new Map<string, Record<string, string[]>>();

    for (const [key, prop] of Object.entries(properties)) {
      const grouped = prop["x-tab-group"];
      if (!grouped) continue;
      if (!isStepMatch(currentSchema, key, currentStep)) continue;

      const filteredOptions: Record<string, string[]> = {};
      const groupedEntries = Object.entries(grouped) as Array<
        [string, string[]]
      >;
      for (const [optionValue, childKeys] of groupedEntries) {
        filteredOptions[optionValue] = childKeys.filter((childKey) => {
          const childProp = properties[childKey];
          if (!childProp) return false;
          return isStepMatch(currentSchema, childKey, currentStep);
        });
      }
      map.set(key, filteredOptions);
    }

    return map;
  }

  function getGroupedFieldsForOption(
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    if (!schema) return [];
    const properties = schema.properties ?? {};
    const childKeys =
      groupedFields.get(controllerKey)?.[String(optionValue)] ?? [];
    const values = { ...$form, [controllerKey]: optionValue };

    return childKeys
      .map<
        [string, JSONSchemaField | undefined]
      >((childKey) => [childKey, properties[childKey]])
      .filter(
        (entry): entry is [string, JSONSchemaField] =>
          Boolean(entry[1]) && isVisibleForValues(schema, entry[0], values),
      );
  }

  function getTabFieldsForOption(
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    if (!schema) return [];
    const properties = schema.properties ?? {};
    const childKeys =
      tabGroupedFields.get(controllerKey)?.[String(optionValue)] ?? [];
    const values = { ...$form, [controllerKey]: optionValue };

    return childKeys
      .map<
        [string, JSONSchemaField | undefined]
      >((childKey) => [childKey, properties[childKey]])
      .filter(
        (entry): entry is [string, JSONSchemaField] =>
          Boolean(entry[1]) && isVisibleForValues(schema, entry[0], values),
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

  function tabOptions(prop: JSONSchemaField) {
    return (
      prop.enum?.map((value, idx) => ({
        value: String(value),
        label: prop["x-enum-labels"]?.[idx] ?? String(value),
      })) ?? []
    );
  }

  function isRequired(key: string) {
    return requiredFields.has(key);
  }
</script>

{#if schema}
  {#each renderOrder as [key, prop] (key)}
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
            {#if groupedFields.get(key)}
              {#each getGroupedFieldsForOption(key, option.value) as [childKey, childProp] (childKey)}
                <div class="py-1.5 first:pt-0 last:pb-0">
                  {#if isTabsEnum(childProp)}
                    {@const childOptions = tabOptions(childProp)}
                    {#if childProp.title}
                      <div class="text-sm font-medium mb-3">
                        {childProp.title}
                      </div>
                    {/if}
                    <Tabs
                      bind:value={$form[childKey]}
                      options={childOptions}
                      disableMarginTop
                    >
                      {#each childOptions as childOption}
                        <TabsContent value={childOption.value}>
                          {#if tabGroupedFields.get(childKey)}
                            {#each getTabFieldsForOption(childKey, childOption.value) as [tabKey, tabProp]}
                              <div class="py-1.5 first:pt-0 last:pb-0">
                                <SchemaField
                                  id={tabKey}
                                  prop={tabProp}
                                  optional={!isRequired(tabKey)}
                                  errors={errors?.[tabKey]}
                                  bind:value={$form[tabKey]}
                                  bind:checked={$form[tabKey]}
                                  {onStringInputChange}
                                  {handleFileUpload}
                                  options={isRadioEnum(tabProp)
                                    ? radioOptions(tabProp)
                                    : undefined}
                                  name={`${tabKey}-radio`}
                                />
                              </div>
                            {/each}
                          {/if}
                        </TabsContent>
                      {/each}
                    </Tabs>
                  {:else}
                    <SchemaField
                      id={childKey}
                      prop={childProp}
                      optional={!isRequired(childKey)}
                      errors={errors?.[childKey]}
                      bind:value={$form[childKey]}
                      bind:checked={$form[childKey]}
                      {onStringInputChange}
                      {handleFileUpload}
                      options={isRadioEnum(childProp)
                        ? radioOptions(childProp)
                        : undefined}
                      name={`${childKey}-radio`}
                    />
                  {/if}
                </div>
              {/each}
            {/if}
          </svelte:fragment>
        </Radio>
      </div>
    {:else if isTabsEnum(prop)}
      {@const options = tabOptions(prop)}
      <div class="py-1.5 first:pt-0 last:pb-0">
        {#if prop.title}
          <div class="text-sm font-medium mb-3">{prop.title}</div>
        {/if}
        <Tabs bind:value={$form[key]} options={options} disableMarginTop>
          {#each options as option}
            <TabsContent value={option.value}>
              {#if tabGroupedFields.get(key)}
                {#each getTabFieldsForOption(key, option.value) as [childKey, childProp]}
                  <div class="py-1.5 first:pt-0 last:pb-0">
                    <SchemaField
                      id={childKey}
                      prop={childProp}
                      optional={!isRequired(childKey)}
                      errors={errors?.[childKey]}
                      bind:value={$form[childKey]}
                      bind:checked={$form[childKey]}
                      {onStringInputChange}
                      {handleFileUpload}
                      options={isRadioEnum(childProp)
                        ? radioOptions(childProp)
                        : undefined}
                      name={`${childKey}-radio`}
                    />
                  </div>
                {/each}
              {/if}
            </TabsContent>
          {/each}
        </Tabs>
      </div>
    {:else}
      <div class="py-1.5 first:pt-0 last:pb-0">
        <SchemaField
          id={key}
          {prop}
          optional={!isRequired(key)}
          errors={errors?.[key]}
          bind:value={$form[key]}
          bind:checked={$form[key]}
          {onStringInputChange}
          {handleFileUpload}
          options={isRadioEnum(prop) ? radioOptions(prop) : undefined}
          name={`${key}-radio`}
        />
      </div>
    {/if}
  {/each}
{/if}
