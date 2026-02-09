<script lang="ts">
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import SchemaField from "./SchemaField.svelte";
  import ConnectionTypeSelector from "./ConnectionTypeSelector.svelte";
  import GroupedFieldsRenderer from "./GroupedFieldsRenderer.svelte";
  import type { JSONSchemaField, MultiStepFormSchema } from "./schemas/types";
  import {
    buildEnumOptions,
    getConditionalValues,
    isDisabledForValues,
    isEnumWithDisplay,
    isRadioEnum,
    isSelectEnum,
    isStepMatch,
    isTabsEnum,
    isVisibleForValues,
    radioOptions,
    selectOptions,
    tabOptions,
  } from "./schema-utils";
  import type { ComponentType, SvelteComponent } from "svelte";

  // Icon mapping for select options
  export let iconMap: Record<string, ComponentType<SvelteComponent>> = {};

  // Use `any` for form values since field types are determined by JSON schema at runtime
  type FormData = Record<string, any>;

  // Superforms-compatible store type with optional taint control
  type SuperFormStore = {
    subscribe: (run: (value: FormData) => void) => () => void;
    update: (
      updater: (value: FormData) => FormData,
      options?: { taint?: boolean },
    ) => void;
  };

  export let schema: MultiStepFormSchema | null = null;
  export let step: string | undefined = undefined;
  export let form: SuperFormStore;
  // Use `any` to be compatible with superforms' complex ValidationErrors type
  export let errors: any;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  $: stepFilter = step;
  $: groupedFields = schema
    ? buildGroupedFields(schema, stepFilter)
    : new Map<string, Record<string, string[]>>();
  $: tabGroupedFields = schema
    ? buildTabGroupedFields(schema, stepFilter)
    : new Map<string, Record<string, string[]>>();
  $: groupedChildKeys = new Set([
    ...Array.from(groupedFields.values()).flatMap((group) =>
      Object.values(group).flat(),
    ),
    ...Array.from(tabGroupedFields.values()).flatMap((group) =>
      Object.values(group).flat(),
    ),
  ]);
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
          } else if (isUnset && isSelectEnum(prop) && prop.enum?.length) {
            $form[key] = String(prop.enum[0]);
          } else if (isUnset && isRichSelectEnum(prop) && prop.enum?.length) {
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

  // Apply const values and conditional defaults from allOf/if/then branches.
  // This ensures that schema-defined constraints are always enforced.
  $: if (schema) {
    const currentValues = $form;
    const conditionalValues = getConditionalValues(schema, currentValues);

    // Check if any conditional values differ from current form values
    const needsUpdate = Object.entries(conditionalValues).some(
      ([key, value]) => {
        if (!isStepMatch(schema, key, stepFilter)) return false;
        return currentValues[key] !== value;
      },
    );

    if (needsUpdate) {
      form.update(
        ($form) => {
          for (const [key, value] of Object.entries(conditionalValues)) {
            if (!isStepMatch(schema, key, stepFilter)) continue;
            $form[key] = value;
          }
          return $form;
        },
        { taint: false },
      );
    }
  }

  function isRichSelectEnum(prop: JSONSchemaField) {
    return Boolean(
      prop.enum &&
        prop["x-display"] === "select" &&
        prop["x-select-style"] === "rich",
    );
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

  function buildFieldGroups(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
    groupKey: "x-grouped-fields" | "x-tab-group",
  ): Map<string, Record<string, string[]>> {
    const properties = currentSchema.properties ?? {};
    const map = new Map<string, Record<string, string[]>>();

    for (const [key, prop] of Object.entries(properties)) {
      const grouped = prop[groupKey];
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

  function buildGroupedFields(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
  ): Map<string, Record<string, string[]>> {
    return buildFieldGroups(currentSchema, currentStep, "x-grouped-fields");
  }

  function buildTabGroupedFields(
    currentSchema: MultiStepFormSchema,
    currentStep: string | undefined,
  ): Map<string, Record<string, string[]>> {
    return buildFieldGroups(currentSchema, currentStep, "x-tab-group");
  }

  function getFieldsForOption(
    fieldMap: Map<string, Record<string, string[]>>,
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    if (!schema) return [];
    const properties = schema.properties ?? {};
    const childKeys = fieldMap.get(controllerKey)?.[String(optionValue)] ?? [];
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

  function getGroupedFieldsForOption(
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    return getFieldsForOption(groupedFields, controllerKey, optionValue);
  }

  function getTabFieldsForOption(
    controllerKey: string,
    optionValue: string | number | boolean,
  ) {
    return getFieldsForOption(tabGroupedFields, controllerKey, optionValue);
  }

  // Local wrapper to pass component's iconMap to imported selectOptions
  function getSelectOptions(prop: JSONSchemaField) {
    return selectOptions(prop, iconMap);
  }

  // Local wrapper to pass iconMap to buildEnumOptions (for GroupedFieldsRenderer)
  function buildEnumOptionsWithIconMap(
    prop: JSONSchemaField,
    includeDescription: boolean,
    includeIcons: boolean = false,
  ) {
    return buildEnumOptions(prop, {
      includeDescription,
      includeIcons,
      iconMap,
    });
  }

  function isRequired(key: string) {
    return requiredFields.has(key);
  }

  function isDisabled(key: string) {
    if (!schema) return false;
    return isDisabledForValues(schema, key, $form);
  }

  function handleSelectChange(key: string, newValue: string) {
    if (!schema) return;
    const prop = schema.properties?.[key];
    if (!prop) return;

    // Clear all grouped fields when select value changes
    const groupedFieldsMap = prop["x-grouped-fields"];
    if (groupedFieldsMap) {
      form.update(
        ($form) => {
          // Get all child keys from all groups
          const allChildKeys = new Set(
            Object.values(groupedFieldsMap).flat() as string[],
          );

          // Also collect keys from tab groups of the grouped fields
          for (const childKey of allChildKeys) {
            const childProp = schema.properties?.[childKey];
            const tabGroup = childProp?.["x-tab-group"];
            if (tabGroup) {
              const tabKeys = Object.values(tabGroup).flat() as string[];
              tabKeys.forEach((k) => allChildKeys.add(k));
            }
          }

          // Clear all child keys to empty first (including nested tab fields)
          for (const childKey of allChildKeys) {
            const childProp = schema.properties?.[childKey];
            if (childProp?.["x-ui-only"]) continue; // Don't clear UI-only fields
            $form[childKey] = "";
          }

          // Ensure UI-only enum fields have valid values for conditional matching
          for (const childKey of allChildKeys) {
            const childProp = schema.properties?.[childKey];
            if (!childProp?.["x-ui-only"]) continue;
            // If it's a tabs/select enum, ensure it has a value
            if (childProp.enum?.length && !$form[childKey]) {
              $form[childKey] = childProp.default ?? String(childProp.enum[0]);
            }
          }

          $form[key] = newValue;

          // Apply conditional defaults from allOf/if/then branches
          const conditionalValues = getConditionalValues(schema, $form);
          for (const [condKey, value] of Object.entries(conditionalValues)) {
            $form[condKey] = value;
          }

          // For fields still empty, fall back to base defaults
          for (const childKey of allChildKeys) {
            const childProp = schema.properties?.[childKey];
            if (childProp?.["x-ui-only"]) continue;
            if ($form[childKey] === "" && childProp?.default !== undefined) {
              $form[childKey] = childProp.default;
            }
          }

          return $form;
        },
        { taint: true },
      );
    }
  }
</script>

{#if schema}
  {#each renderOrder as [key, prop] (key)}
    {#if isRichSelectEnum(prop)}
      {@const options = getSelectOptions(prop)}
      <div class="py-1.5 first:pt-0 last:pb-0">
        <ConnectionTypeSelector
          bind:value={$form[key]}
          {options}
          label={prop.title ?? ""}
          onChange={(newValue) => handleSelectChange(key, newValue)}
        />
        {#if groupedFields.get(key)}
          <GroupedFieldsRenderer
            fields={getGroupedFieldsForOption(key, $form[key])}
            formStore={form}
            {errors}
            {onStringInputChange}
            {handleFileUpload}
            {isRequired}
            {isDisabled}
            {getTabFieldsForOption}
            {tabGroupedFields}
            buildEnumOptions={buildEnumOptionsWithIconMap}
          />
        {/if}
      </div>
    {:else if isSelectEnum(prop)}
      {@const options = getSelectOptions(prop)}
      <div class="py-1.5 first:pt-0 last:pb-0">
        <Select
          id={key}
          bind:value={$form[key]}
          {options}
          label={prop.title ?? ""}
          placeholder={prop["x-placeholder"] ?? "Select an option"}
          tooltip={prop.description ?? ""}
          optional={!isRequired(key)}
          full
          onChange={(newValue) => handleSelectChange(key, newValue)}
        />
        {#if groupedFields.get(key)}
          <GroupedFieldsRenderer
            fields={getGroupedFieldsForOption(key, $form[key])}
            formStore={form}
            {errors}
            {onStringInputChange}
            {handleFileUpload}
            {isRequired}
            {isDisabled}
            {getTabFieldsForOption}
            {tabGroupedFields}
            buildEnumOptions={buildEnumOptionsWithIconMap}
          />
        {/if}
      </div>
    {:else if isRadioEnum(prop)}
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
              <GroupedFieldsRenderer
                fields={getGroupedFieldsForOption(key, option.value)}
                formStore={form}
                {errors}
                {onStringInputChange}
                {handleFileUpload}
                {isRequired}
                {isDisabled}
                {getTabFieldsForOption}
                {tabGroupedFields}
                buildEnumOptions={buildEnumOptionsWithIconMap}
              />
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
        <Tabs bind:value={$form[key]} {options} disableMarginTop>
          {#each options as option (option.value)}
            <TabsContent value={option.value}>
              {#if tabGroupedFields.get(key)}
                {#each getTabFieldsForOption(key, option.value) as [childKey, childProp] (childKey)}
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
                      disabled={isDisabled(childKey)}
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
          disabled={isDisabled(key)}
        />
      </div>
    {/if}
  {/each}
{/if}
