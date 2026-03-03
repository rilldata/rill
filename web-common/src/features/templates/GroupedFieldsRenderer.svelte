<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
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

  function getSelectOptions(prop: JSONSchemaField) {
    return buildEnumOptions(prop, true, true);
  }
</script>

{#each fields as [childKey, childProp] (childKey)}
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
              {#each getTabFieldsForOption(childKey, childOption.value) as [tabKey, tabProp] (tabKey)}
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
