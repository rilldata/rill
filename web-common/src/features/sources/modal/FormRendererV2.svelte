<script lang="ts">
  import ParametersForm from "./ParametersForm.svelte";
  import DSNForm from "./DSNForm.svelte";
  import ClickHouseForm from "./ClickHouseForm.svelte";
  import Tabs from "@rilldata/web-common/components/forms/Tabs.svelte";
  import { TabsContent } from "@rilldata/web-common/components/tabs";
  import { CONNECTION_TAB_OPTIONS } from "./constants";
  import type { ConnectorType } from "./types";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import type { AddDataFormType } from "./types";
  import type { ConnectorHandler } from "./connector-handlers";
  import type { SuperFormState } from "./form-factory";
  import type { ClickHouseConnectorType } from "./constants";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let connectionTab: ConnectorType;
  export let connectorHandler: ConnectorHandler;
  export let forms: {
    paramsForm: SuperFormState | null;
    dsnForm: SuperFormState | null;
  } | null;
  export let onStringInputChange: (event: Event) => void;

  // ClickHouse-specific props
  export let clickhouseConnectorType: ClickHouseConnectorType;
  export let clickhouseForms: {
    paramsForm: SuperFormState;
    dsnForm: SuperFormState;
  } | null;
  export let onClickHouseStringInputChange: (event: Event) => void;

  // Computed properties
  $: hasDsnFormOption = connectorHandler.hasDsnFormOption(connector);
  $: hasOnlyDsn = connectorHandler.hasOnlyDsn(connector);
  $: isClickhouseConnector = connector.name === "clickhouse";

  // Get properties for current form
  $: paramsProperties = connectorHandler.getFilteredProperties(
    connector,
    formType,
  );
  $: dsnProperties = connectorHandler.getDsnProperties(connector);
</script>

{#if isClickhouseConnector && clickhouseForms}
  <ClickHouseForm
    bind:connectorType={clickhouseConnectorType}
    bind:connectionTab
    paramsForm={clickhouseForms.paramsForm.form}
    dsnForm={clickhouseForms.dsnForm.form}
    paramsErrors={clickhouseForms.paramsForm.errors}
    dsnErrors={clickhouseForms.dsnForm.errors}
    paramsEnhance={clickhouseForms.paramsForm.enhance}
    dsnEnhance={clickhouseForms.dsnForm.enhance}
    paramsSubmit={clickhouseForms.paramsForm.submit}
    dsnSubmit={clickhouseForms.dsnForm.submit}
    paramsFormId={connectorHandler.getFormId(connector, "params")}
    dsnFormId={connectorHandler.getFormId(connector, "dsn")}
    {dsnProperties}
    filteredProperties={paramsProperties}
    onStringInputChange={onClickHouseStringInputChange}
  />
{:else if hasDsnFormOption && forms}
  <Tabs
    bind:value={connectionTab}
    options={CONNECTION_TAB_OPTIONS}
    disableMarginTop
  >
    <TabsContent value="parameters">
      {#if forms.paramsForm}
        <ParametersForm
          properties={paramsProperties}
          formId={connectorHandler.getFormId(connector, "params")}
          form={forms.paramsForm.form}
          errors={forms.paramsForm.errors}
          enhance={forms.paramsForm.enhance}
          submit={forms.paramsForm.submit}
          {onStringInputChange}
        />
      {/if}
    </TabsContent>
    <TabsContent value="dsn">
      {#if forms.dsnForm}
        <DSNForm
          properties={dsnProperties}
          formId={connectorHandler.getFormId(connector, "dsn")}
          form={forms.dsnForm.form}
          errors={forms.dsnForm.errors}
          enhance={forms.dsnForm.enhance}
          submit={forms.dsnForm.submit}
        />
      {/if}
    </TabsContent>
  </Tabs>
{:else if formType === "connector" && hasOnlyDsn && forms}
  <!-- Connector with only DSN - show DSN form directly -->
  {#if forms.dsnForm}
    <DSNForm
      properties={dsnProperties}
      formId={connectorHandler.getFormId(connector, "dsn")}
      form={forms.dsnForm.form}
      errors={forms.dsnForm.errors}
      enhance={forms.dsnForm.enhance}
      submit={forms.dsnForm.submit}
    />
  {/if}
{:else if forms}
  <!-- Default parameters form -->
  {#if forms.paramsForm}
    <ParametersForm
      properties={paramsProperties}
      formId={connectorHandler.getFormId(connector, "params")}
      form={forms.paramsForm.form}
      errors={forms.paramsForm.errors}
      enhance={forms.paramsForm.enhance}
      submit={forms.paramsForm.submit}
      {onStringInputChange}
    />
  {/if}
{/if}
