<script lang="ts">
  import ParametersForm from "./ParametersForm.svelte";
  import DSNForm from "./DSNForm.svelte";
  import ClickHouseForm from "./ClickHouseForm.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { CONNECTION_TAB_OPTIONS } from "./constants";
  import type { ConnectorType } from "./types";

  export let connector: any;
  export let formType: any;
  export let connectionTab: ConnectorType;
  export let hasDsnFormOption: boolean;
  export let paramsFormId: string;
  export let dsnFormId: string;
  export let paramsForm: any;
  export let dsnForm: any;
  export let paramsErrors: any;
  export let dsnErrors: any;
  export let paramsEnhance: any;
  export let dsnEnhance: any;
  export let paramsSubmit: any;
  export let dsnSubmit: any;
  export let filteredParamsProperties: any[];
  export let filteredDsnProperties: any[];
  export let onStringInputChange: (event: Event) => void;

  // ClickHouse-specific props
  export let clickhouseConnectorType: any;
  export let clickhouseParamsForm: any;
  export let clickhouseDsnForm: any;
  export let clickhouseParamsErrors: any;
  export let clickhouseDsnErrors: any;
  export let clickhouseParamsEnhance: any;
  export let clickhouseDsnEnhance: any;
  export let clickhouseParamsSubmit: any;
  export let clickhouseDsnSubmit: any;
  export let clickhouseParamsFormId: string;
  export let clickhouseDsnFormId: string;
  export let clickhouseDsnProperties: any[];
  export let clickhouseFilteredProperties: any[];
  export let onClickHouseStringInputChange: (event: Event) => void;
</script>

{#if connector.name === "clickhouse"}
  <ClickHouseForm
    bind:connectorType={clickhouseConnectorType}
    bind:connectionTab
    paramsForm={clickhouseParamsForm}
    dsnForm={clickhouseDsnForm}
    paramsErrors={clickhouseParamsErrors}
    dsnErrors={clickhouseDsnErrors}
    paramsEnhance={clickhouseParamsEnhance}
    dsnEnhance={clickhouseDsnEnhance}
    paramsSubmit={clickhouseParamsSubmit}
    dsnSubmit={clickhouseDsnSubmit}
    paramsFormId={clickhouseParamsFormId}
    dsnFormId={clickhouseDsnFormId}
    dsnProperties={clickhouseDsnProperties}
    filteredProperties={clickhouseFilteredProperties}
    onStringInputChange={onClickHouseStringInputChange}
  />
{:else if hasDsnFormOption}
  <Tabs
    bind:value={connectionTab}
    options={CONNECTION_TAB_OPTIONS}
    disableMarginTop
  >
    <TabsContent value="parameters">
      <ParametersForm
        properties={filteredParamsProperties}
        formId={paramsFormId}
        form={paramsForm}
        errors={paramsErrors}
        enhance={paramsEnhance}
        submit={paramsSubmit}
        {onStringInputChange}
      />
    </TabsContent>
    <TabsContent value="dsn">
      <DSNForm
        properties={filteredDsnProperties}
        formId={dsnFormId}
        form={dsnForm}
        errors={dsnErrors}
        enhance={dsnEnhance}
        submit={dsnSubmit}
      />
    </TabsContent>
  </Tabs>
{:else if formType === "connector" && connector.configProperties?.some((property) => property.key === "dsn")}
  <!-- Connector with only DSN - show DSN form directly -->
  <DSNForm
    properties={filteredDsnProperties}
    formId={dsnFormId}
    form={dsnForm}
    errors={dsnErrors}
    enhance={dsnEnhance}
    submit={dsnSubmit}
  />
{:else}
  <!-- Default parameters form -->
  <ParametersForm
    properties={filteredParamsProperties}
    formId={paramsFormId}
    form={paramsForm}
    errors={paramsErrors}
    enhance={paramsEnhance}
    submit={paramsSubmit}
    {onStringInputChange}
  />
{/if}
