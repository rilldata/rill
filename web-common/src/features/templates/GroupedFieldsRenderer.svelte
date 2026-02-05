<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import SchemaField from "./SchemaField.svelte";
  import type { JSONSchemaField } from "./schemas/types";
  import type { ComponentType, SvelteComponent } from "svelte";

  type FieldEntry = [string, JSONSchemaField];
  type EnumOption = {
    value: string;
    label: string;
    description?: string;
    icon?: ComponentType<SvelteComponent>;
  };

  export let fields: FieldEntry[];
  export let form: Record<string, any>;
  export let errors: any;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;
  export let isRequired: (key: string) => boolean;
  export let isDisabled: (key: string) => boolean;
  export let getTabFieldsForOption: (
    controllerKey: string,
    optionValue: string | number | boolean,
  ) => FieldEntry[];
  export let tabGroupedFields: Map<string, Record<string, string[]>>;
  export let buildEnumOptions: (
    prop: JSONSchemaField,
    includeDescription: boolean,
    includeIcons?: boolean,
  ) => EnumOption[];

  function isEnumWithDisplay(
    prop: JSONSchemaField,
    displayType: "radio" | "tabs" | "select",
  ) {
    return Boolean(prop.enum && prop["x-display"] === displayType);
  }

  function isRadioEnum(prop: JSONSchemaField) {
    return isEnumWithDisplay(prop, "radio");
  }

  function isTabsEnum(prop: JSONSchemaField) {
    return isEnumWithDisplay(prop, "tabs");
  }

  function isSelectEnum(prop: JSONSchemaField) {
    return isEnumWithDisplay(prop, "select");
  }

  function tabOptions(prop: JSONSchemaField) {
    return buildEnumOptions(prop, false);
  }

  function selectOptions(prop: JSONSchemaField) {
    return buildEnumOptions(prop, true, true);
  }

  function radioOptions(prop: JSONSchemaField) {
    return buildEnumOptions(prop, true);
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
      <Tabs bind:value={form[childKey]} options={childOptions} disableMarginTop>
        {#each childOptions as childOption (childOption.value)}
          <TabsContent value={childOption.value}>
            {#if tabGroupedFields.get(childKey)}
              {#each getTabFieldsForOption(childKey, childOption.value) as [tabKey, tabProp] (tabKey)}
                <div class="py-1.5 first:pt-0 last:pb-0">
                  {#if isSelectEnum(tabProp)}
                    {@const tabSelectOptions = selectOptions(tabProp)}
                    <Select
                      id={tabKey}
                      bind:value={form[tabKey]}
                      options={tabSelectOptions}
                      label={tabProp.title ?? ""}
                      placeholder={tabProp["x-placeholder"] ?? "Select an option"}
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
                      bind:value={form[tabKey]}
                      bind:checked={form[tabKey]}
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
      {@const childSelectOptions = selectOptions(childProp)}
      <Select
        id={childKey}
        bind:value={form[childKey]}
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
        bind:value={form[childKey]}
        bind:checked={form[childKey]}
        {onStringInputChange}
        {handleFileUpload}
        options={isRadioEnum(childProp) ? radioOptions(childProp) : undefined}
        name={`${childKey}-radio`}
        disabled={isDisabled(childKey)}
      />
    {/if}
  </div>
{/each}
