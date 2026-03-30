<script lang="ts">
  import SourceSelector from "@rilldata/web-common/features/add-data/manager/SourceSelector.svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import SourceForm from "@rilldata/web-common/features/add-data/form/SourceForm.svelte";
  import ImportTableForm from "@rilldata/web-common/features/add-data/form/ImportTableForm.svelte";
  import GenerateDashboardStatus from "@rilldata/web-common/features/add-data/manager/GenerateDashboardStatus.svelte";
  import {
    type AddDataConfig,
    AddDataStep,
    type ImportStepConfig,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ConnectorHeader from "@rilldata/web-common/features/add-data/manager/ConnectorHeader.svelte";
  import ImportDataStatus from "@rilldata/web-common/features/add-data/manager/ImportDataStatus.svelte";
  import {
    AddDataStateManager,
    TransitionEventType,
  } from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";
  import {
    getConnectorDriverForConnector,
    getConnectorDriverForSchema,
  } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  const {
    config,
    initSchema,
    initConnector,
    onDone,
    onClose = () => {},
    onStepChange,
  }: {
    config: AddDataConfig;
    initSchema?: string;
    initConnector?: string;
    onDone?: () => void;
    onClose?: () => void;
    onStepChange?: (step: AddDataStep) => void;
  } = $props();

  const stateManager = $derived(
    new AddDataStateManager(onDone, onClose, onStepChange),
  );
  $effect(() => void init(initConnector, initSchema));

  const runtimeClient = useRuntimeClient();

  let stepState = $derived(stateManager.state);

  let schema = $derived((stepState as any).schema as string | undefined);
  let connector = $derived((stepState as any).connector as string | undefined);

  const SizeClassMap: Partial<Record<AddDataStep, string>> = {
    [AddDataStep.SelectConnector]: "h-fit w-[900px]",
    [AddDataStep.Import]: "h-fit w-[550px]",
  };
  let sizeClass = $derived(
    SizeClassMap[stepState.step] ?? "h-[630px] w-[900px]",
  );
  let shouldShowHeader = $derived(
    stepState.step === AddDataStep.CreateConnector ||
      stepState.step === AddDataStep.CreateModel ||
      stepState.step === AddDataStep.ExploreConnector,
  );

  async function init(connector?: string, schema?: string) {
    let driver: V1ConnectorDriver | undefined = undefined;
    if (connector) {
      const analyzedConnector = await getConnectorDriverForConnector(
        runtimeClient,
        connector,
      );
      driver = analyzedConnector?.driver;
      schema = driver?.name ?? schema;
    } else if (schema) {
      driver = getConnectorDriverForSchema(schema);
    }
    stateManager.transition({
      type: TransitionEventType.Init,
      connector,
      schema,
      driver,
    });
  }

  function schemaSelected(schema: string) {
    const driver = getConnectorDriverForSchema(schema);
    if (!driver) return;
    stateManager.transition({
      type: TransitionEventType.SchemaSelected,
      schema,
      driver,
    });
  }

  async function connectorSelected(
    connector: string,
    connectorFormValues: Record<string, any>,
  ) {
    const analyzedConnector = await getConnectorDriverForConnector(
      runtimeClient,
      connector,
    );
    if (!analyzedConnector?.driver) return;
    stateManager.transition({
      type: TransitionEventType.ConnectorSelected,
      schema: analyzedConnector.driver.name!,
      driver: analyzedConnector.driver,
      connector,
      connectorFormValues,
    });
  }

  function importConfigured(importConfig: ImportStepConfig) {
    stateManager.transition({
      type: TransitionEventType.ImportConfigured,
      config: importConfig,
    });
  }

  function onBack() {
    stateManager.transition({ type: TransitionEventType.Back });
  }
</script>

<div
  class="flex flex-col {sizeClass} bg-surface-background border rounded-lg shadow-sm"
>
  {#if shouldShowHeader && schema}
    <ConnectorHeader
      {config}
      schemaName={schema}
      connectorName={connector}
      onConnectorChange={(newConnector) => connectorSelected(newConnector, {})}
      onNewConnector={() => schemaSelected(schema)}
    />
  {/if}

  {#if stepState.step === AddDataStep.SelectConnector}
    <SourceSelector {config} onSelect={schemaSelected} {onBack} />
  {:else if stepState.step === AddDataStep.CreateConnector}
    <ConnectorForm
      step={stepState}
      onSubmit={(connectorName, connectorFormValues) =>
        void connectorSelected(connectorName, connectorFormValues)}
      {onBack}
      {onClose}
    />
  {:else if stepState.step === AddDataStep.CreateModel}
    {#key stepState.connector}
      <SourceForm
        {config}
        step={stepState}
        onSubmit={importConfigured}
        {onBack}
      />
    {/key}
  {:else if stepState.step === AddDataStep.ExploreConnector}
    {#key stepState.connector}
      <ImportTableForm
        {config}
        step={stepState}
        onSubmit={importConfigured}
        {onBack}
      />
    {/key}
  {:else if stepState.step === AddDataStep.Import}
    {@const isImportOnlyStep =
      stepState.config.importSteps.length === 1 &&
      stepState.config.importSteps[0] === ImportDataStep.CreateModel}
    {#if isImportOnlyStep}
      <!-- Special case for import only, we show additional options to handle success and failures. -->
      <ImportDataStatus importAddDataStep={stepState} {onClose} />
    {:else}
      <GenerateDashboardStatus
        importAddDataStep={stepState}
        {onBack}
        {onClose}
      />
    {/if}
  {/if}
</div>
