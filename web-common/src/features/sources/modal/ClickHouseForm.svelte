<script lang="ts">
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import ConnectorTypeSelector from "@rilldata/web-common/components/forms/ConnectorTypeSelector.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";
  import { CONNECTOR_TYPE_OPTIONS, CONNECTION_TAB_OPTIONS } from "./constants";
  import type { ClickHouseConnectorType } from "./constants";
  import type { ConnectorType } from "./types";
  import { inferSourceName } from "../sourceUtils";

  export let connectorType: ClickHouseConnectorType;
  export let connectionTab: ConnectorType;
  export let paramsForm: any;
  export let dsnForm: any;
  export let paramsErrors: any;
  export let dsnErrors: any;
  export let paramsEnhance: any;
  export let dsnEnhance: any;
  export let paramsSubmit: any;
  export let dsnSubmit: any;
  export let paramsFormId: string;
  export let dsnFormId: string;
  export let dsnProperties: any[];
  export let filteredProperties: any[];
  export let onStringInputChange: (event: Event) => void;

  function handleConnectorTypeChange(event: CustomEvent) {
    connectorType = event.detail.value;
  }

  function handleConnectionTabChange(event: CustomEvent) {
    connectionTab = event.detail.value;
  }
</script>

<div class="h-full w-full flex flex-col">
  <div>
    <ConnectorTypeSelector
      bind:value={connectorType}
      options={CONNECTOR_TYPE_OPTIONS}
      on:change={handleConnectorTypeChange}
    />
    {#if connectorType === "rill-managed"}
      <div class="mt-4">
        <InformationalField
          description="This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance."
        />
      </div>
    {/if}
  </div>

  {#if connectorType === "self-hosted" || connectorType === "clickhouse-cloud"}
    <Tabs bind:value={connectionTab} options={CONNECTION_TAB_OPTIONS}>
      <TabsContent value="parameters">
        <form
          id={paramsFormId}
          class="flex-grow overflow-y-auto"
          use:paramsEnhance
          on:submit|preventDefault={paramsSubmit}
        >
          {#each filteredProperties as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            {@const isPortField = propertyKey === "port"}
            {@const isSSLField = propertyKey === "ssl"}

            <div class="py-1.5 first:pt-0 last:pb-0">
              {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
                <Input
                  id={propertyKey}
                  label={property.displayName}
                  placeholder={property.placeholder}
                  optional={!property.required}
                  secret={property.secret}
                  hint={property.hint}
                  errors={normalizeErrors(paramsErrors[propertyKey])}
                  bind:value={$paramsForm[propertyKey]}
                  onInput={(_, e) => onStringInputChange(e)}
                  alwaysShowError
                  disabled={connectorType === "clickhouse-cloud" && isPortField}
                />
              {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
                <Checkbox
                  id={propertyKey}
                  bind:checked={$paramsForm[propertyKey]}
                  label={property.displayName}
                  hint={property.hint}
                  optional={connectorType === "clickhouse-cloud" && isSSLField
                    ? false
                    : !property.required}
                  disabled={connectorType === "clickhouse-cloud" && isSSLField}
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
        </form>
      </TabsContent>
      <TabsContent value="dsn">
        <form
          id={dsnFormId}
          class="flex-grow overflow-y-auto"
          use:dsnEnhance
          on:submit|preventDefault={dsnSubmit}
        >
          {#each dsnProperties as property (property.key)}
            {@const propertyKey = property.key ?? ""}
            <div class="py-1.0 first:pt-0 last:pb-0">
              <Input
                id={propertyKey}
                label={property.displayName}
                placeholder={property.placeholder}
                secret={property.secret}
                hint={property.hint}
                errors={normalizeErrors(dsnErrors[propertyKey])}
                bind:value={$dsnForm[propertyKey]}
                alwaysShowError
              />
            </div>
          {/each}
        </form>
      </TabsContent>
    </Tabs>
  {/if}
</div>
