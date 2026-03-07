<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { slide } from "svelte/transition";
  import SchemaField from "./SchemaField.svelte";
  import type { JSONSchemaField } from "./schemas/types";
  import {
    type EnumOption,
    isRadioEnum,
    isSelectEnum,
    isTabsEnum,
    radioOptions,
    tabOptions,
  } from "./schema-utils";

  type FieldEntry = [string, JSONSchemaField];

  // Superforms-compatible store type (matches JSONSchemaFormRenderer)
  type FormStore = {
    subscribe: (run: (value: Record<string, any>) => void) => () => void;
    update: (
      updater: (value: Record<string, any>) => Record<string, any>,
      options?: { taint?: boolean },
    ) => void;
  };

  export let fields: FieldEntry[];
  export let formStore: FormStore;
  export let errors: any;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (
    file: File,
    fieldKey: string,
  ) => Promise<string>;
  export let isRequired: (key: string) => boolean;
  export let isDisabled: (key: string) => boolean;
  export let getTabFieldsForOption: (
    controllerKey: string,
    optionValue: string | number | boolean,
  ) => FieldEntry[];
  export let tabGroupedFields: Map<string, Record<string, string[]>>;
  // Passed from parent to include iconMap context
  export let buildEnumOptions: (
    prop: JSONSchemaField,
    includeDescription: boolean,
    includeIcons?: boolean,
  ) => EnumOption[];

  $: regularFields = fields.filter(([, prop]) => !prop["x-advanced"]);
  $: advancedFields = fields.filter(([, prop]) => prop["x-advanced"]);
  $: hasTabAdvanced = fields.some(([childKey, childProp]) => {
    if (!isTabsEnum(childProp)) return false;
    if (!tabGroupedFields.get(childKey)) return false;
    const currentValue = $formStore[childKey];
    const tabFields = getTabFieldsForOption(childKey, currentValue);
    return tabFields.some(([, p]) => p["x-advanced"]);
  });
  $: hasAnyAdvanced = advancedFields.length > 0 || hasTabAdvanced;

  let showAdvanced = false;

  function getSelectOptions(prop: JSONSchemaField) {
    return buildEnumOptions(prop, true, true);
  }
</script>

{#each regularFields as [childKey, childProp] (childKey)}
  <div class="py-1.5 first:pt-0 last:pb-0">
    {#if isTabsEnum(childProp)}
      {@const childOptions = tabOptions(childProp)}
      {#if childProp.title}
        <div class="text-sm font-medium mb-3">
          {childProp.title}
        </div>
      {/if}
      <Tabs
        bind:value={$formStore[childKey]}
        options={childOptions}
        disableMarginTop
      >
        {#each childOptions as childOption (childOption.value)}
          <TabsContent value={childOption.value}>
            {#if tabGroupedFields.get(childKey)}
              {@const allTabFields = getTabFieldsForOption(
                childKey,
                childOption.value,
              )}
              {@const regularTabFields = allTabFields.filter(
                ([, p]) => !p["x-advanced"],
              )}
              {@const advancedTabFields = allTabFields.filter(
                ([, p]) => p["x-advanced"],
              )}
              {#each regularTabFields as [tabKey, tabProp] (tabKey)}
                <div class="py-1.5 first:pt-0 last:pb-0">
                  {#if isSelectEnum(tabProp)}
                    {@const tabSelectOptions = getSelectOptions(tabProp)}
                    <Select
                      id={tabKey}
                      bind:value={$formStore[tabKey]}
                      options={tabSelectOptions}
                      label={tabProp.title ?? ""}
                      placeholder={tabProp["x-placeholder"] ??
                        "Select an option"}
                      tooltip={tabProp.description ?? ""}
                      optional={!isRequired(tabKey)}
                      full
                      disabled={isDisabled(tabKey)}
                    />
                  {:else}
                    <SchemaField
                      id={tabKey}
                      prop={tabProp}
                      optional={!isRequired(tabKey)}
                      errors={errors?.[tabKey]}
                      bind:value={$formStore[tabKey]}
                      bind:checked={$formStore[tabKey]}
                      {onStringInputChange}
                      {handleFileUpload}
                      options={isRadioEnum(tabProp)
                        ? radioOptions(tabProp)
                        : undefined}
                      name={`${tabKey}-radio`}
                      disabled={isDisabled(tabKey)}
                    />
                  {/if}
                </div>
              {/each}
              {#if advancedTabFields.length > 0}
                {#if showAdvanced}
                  {#each advancedTabFields as [tabKey, tabProp] (tabKey)}
                    <div class="py-1.5 first:pt-0 last:pb-0">
                      {#if isSelectEnum(tabProp)}
                        {@const tabSelectOptions = getSelectOptions(tabProp)}
                        <Select
                          id={tabKey}
                          bind:value={$formStore[tabKey]}
                          options={tabSelectOptions}
                          label={tabProp.title ?? ""}
                          placeholder={tabProp["x-placeholder"] ??
                            "Select an option"}
                          tooltip={tabProp.description ?? ""}
                          optional={!isRequired(tabKey)}
                          full
                          disabled={isDisabled(tabKey)}
                        />
                      {:else}
                        <SchemaField
                          id={tabKey}
                          prop={tabProp}
                          optional={!isRequired(tabKey)}
                          errors={errors?.[tabKey]}
                          bind:value={$formStore[tabKey]}
                          bind:checked={$formStore[tabKey]}
                          {onStringInputChange}
                          {handleFileUpload}
                          options={isRadioEnum(tabProp)
                            ? radioOptions(tabProp)
                            : undefined}
                          name={`${tabKey}-radio`}
                          disabled={isDisabled(tabKey)}
                        />
                      {/if}
                    </div>
                  {/each}
                {/if}
              {/if}
            {/if}
          </TabsContent>
        {/each}
      </Tabs>
    {:else if isSelectEnum(childProp)}
      {@const childSelectOptions = getSelectOptions(childProp)}
      <Select
        id={childKey}
        bind:value={$formStore[childKey]}
        options={childSelectOptions}
        label={childProp.title ?? ""}
        placeholder={childProp["x-placeholder"] ?? "Select an option"}
        tooltip={childProp.description ?? ""}
        optional={!isRequired(childKey)}
        full
        disabled={isDisabled(childKey)}
      />
    {:else}
      <SchemaField
        id={childKey}
        prop={childProp}
        optional={!isRequired(childKey)}
        errors={errors?.[childKey]}
        bind:value={$formStore[childKey]}
        bind:checked={$formStore[childKey]}
        {onStringInputChange}
        {handleFileUpload}
        options={isRadioEnum(childProp) ? radioOptions(childProp) : undefined}
        name={`${childKey}-radio`}
        disabled={isDisabled(childKey)}
      />
    {/if}
  </div>
{/each}

{#if hasAnyAdvanced}
  <div class="pt-2">
    <button
      type="button"
      class="advanced-toggle"
      on:click={() => (showAdvanced = !showAdvanced)}
    >
      <svg
        class="chevron"
        class:open={showAdvanced}
        width="12"
        height="12"
        viewBox="0 0 12 12"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M4.5 2.5L8 6L4.5 9.5"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
      Advanced options
    </button>
    {#if showAdvanced}
      <div class="pt-1" transition:slide={{ duration: 150 }}>
        {#each advancedFields as [childKey, childProp] (childKey)}
          <div class="py-1.5 first:pt-0 last:pb-0">
            {#if isSelectEnum(childProp)}
              {@const childSelectOptions = getSelectOptions(childProp)}
              <Select
                id={childKey}
                bind:value={$formStore[childKey]}
                options={childSelectOptions}
                label={childProp.title ?? ""}
                placeholder={childProp["x-placeholder"] ?? "Select an option"}
                tooltip={childProp.description ?? ""}
                optional={!isRequired(childKey)}
                full
                disabled={isDisabled(childKey)}
              />
            {:else}
              <SchemaField
                id={childKey}
                prop={childProp}
                optional={!isRequired(childKey)}
                errors={errors?.[childKey]}
                bind:value={$formStore[childKey]}
                bind:checked={$formStore[childKey]}
                {onStringInputChange}
                {handleFileUpload}
                options={isRadioEnum(childProp)
                  ? radioOptions(childProp)
                  : undefined}
                name={`${childKey}-radio`}
                disabled={isDisabled(childKey)}
              />
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}

<style>
  .advanced-toggle {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    font-weight: 500;
    color: var(--color-text-secondary, #6b7280);
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
  }
  .advanced-toggle:hover {
    color: var(--color-text-primary, #374151);
  }
  .chevron {
    transition: transform 150ms ease;
  }
  .chevron.open {
    transform: rotate(90deg);
  }
</style>
