<script lang="ts">
  import ConnectorTypeSelector from "@rilldata/web-common/components/forms/ConnectorTypeSelector.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { normalizeErrors } from "../../templates/error-utils";
  import {
    CONNECTOR_TYPE_OPTIONS,
    CONNECTION_TAB_OPTIONS,
    type ClickHouseConnectorType,
  } from "./constants";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import AddDataFormSection from "./AddDataFormSection.svelte";
  import type { ClickhouseUiState } from "./AddDataFormManager";
  import type { ConnectorType } from "./types";

  export let clickhouseConnectorType: ClickHouseConnectorType;
  export let clickhouseUiState: ClickhouseUiState | null = null;
  export let connectionTab: ConnectorType;

  export let paramsFormId: string;
  export let paramsEnhance: any;
  export let paramsSubmit: any;
  export let paramsErrors: Record<string, any>;
  export let paramsFormStore: any;

  export let dsnFormId: string;
  export let dsnEnhance: any;
  export let dsnSubmit: any;
  export let dsnErrors: Record<string, any>;
  export let dsnFormStore: any;

  export let onStringInputChange: (event: Event) => void;
</script>

<div class="h-full w-full flex flex-col">
  <div>
    <ConnectorTypeSelector
      bind:value={clickhouseConnectorType}
      options={CONNECTOR_TYPE_OPTIONS}
    />
    {#if clickhouseConnectorType === "rill-managed"}
      <div class="mt-4">
        <InformationalField
          description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
        />
      </div>
    {/if}
  </div>

  {#if clickhouseConnectorType === "self-hosted" || clickhouseConnectorType === "clickhouse-cloud"}
    <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
      <TabsContent value="parameters">
        <AddDataFormSection
          id={paramsFormId}
          enhance={paramsEnhance}
          onSubmit={paramsSubmit}
        >
          {#each clickhouseUiState?.filteredProperties ?? [] as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            {@const isPortField = propertyKey === "port"}
            {@const isSSLField = propertyKey === "ssl"}

            <div class="py-1.5 first:pt-0 last:pb-0">
              {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
                <Input
                  id={propertyKey}
                  label={property.displayName}
                  placeholder={property.placeholder}
                  optional={property.required}
                  secret={property.secret}
                  hint={property.hint}
                  errors={normalizeErrors(paramsErrors?.[propertyKey])}
                  bind:value={$paramsFormStore[propertyKey]}
                  onInput={(_, e) => onStringInputChange(e)}
                  alwaysShowError
                  options={clickhouseConnectorType === "clickhouse-cloud" &&
                  isPortField
                    ? [
                        { value: "8443", label: "8443 (HTTPS)" },
                        { value: "9440", label: "9440 (Native Secure)" },
                      ]
                    : undefined}
                />
              {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                <Checkbox
                  id={propertyKey}
                  bind:checked={$paramsFormStore[propertyKey]}
                  label={property.displayName}
                  hint={property.hint}
                  optional={clickhouseConnectorType === "clickhouse-cloud" &&
                  isSSLField
                    ? false
                    : !property.required}
                  disabled={clickhouseConnectorType === "clickhouse-cloud" &&
                    isSSLField}
                />
              {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
                <InformationalField
                  description={property.description}
                  hint={property.hint}
                  href={property.docsUrl}
                />
              {/if}
            </div>
          {/each}
        </AddDataFormSection>
      </TabsContent>
      <TabsContent value="dsn">
        <AddDataFormSection
          id={dsnFormId}
          enhance={dsnEnhance}
          onSubmit={dsnSubmit}
        >
          {#each clickhouseUiState?.dsnProperties ?? [] as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            <div class="py-1.0 first:pt-0 last:pb-0">
              <Input
                id={propertyKey}
                label={property.displayName}
                placeholder={property.placeholder}
                secret={property.secret}
                hint={property.hint}
                errors={normalizeErrors(dsnErrors?.[propertyKey])}
                bind:value={$dsnFormStore[propertyKey]}
                alwaysShowError
              />
            </div>
          {/each}
        </AddDataFormSection>
      </TabsContent>
    </Tabs>
  {:else}
    <AddDataFormSection
      id={paramsFormId}
      enhance={paramsEnhance}
      onSubmit={paramsSubmit}
    >
      {#each clickhouseUiState?.filteredProperties ?? [] as property (property.key)}
        {@const propertyKey = property.key ?? ""}
        <div class="py-1.5 first:pt-0 last:pb-0">
          {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
            <Input
              id={propertyKey}
              label={property.displayName}
              placeholder={property.placeholder}
              optional={!property.required}
              secret={property.secret}
              hint={property.hint}
              errors={normalizeErrors(paramsErrors?.[propertyKey])}
              bind:value={$paramsFormStore[propertyKey]}
              onInput={(_, e) => onStringInputChange(e)}
              alwaysShowError
            />
          {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
            <Checkbox
              id={propertyKey}
              bind:checked={$paramsFormStore[propertyKey]}
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
      {/each}
    </AddDataFormSection>
  {/if}
</div>
