<script lang="ts">
  import { page } from "$app/stores";
  import SourceSelector from "@rilldata/web-common/features/add-data/manager/SourceSelector.svelte";
  import ConnectorForm from "@rilldata/web-common/features/add-data/form/ConnectorForm.svelte";
  import SourceForm from "@rilldata/web-common/features/add-data/form/SourceForm.svelte";
  import ImportTableForm from "@rilldata/web-common/features/add-data/form/ImportTableForm.svelte";
  import GenerateDashboardStatus from "@rilldata/web-common/features/add-data/manager/GenerateDashboardStatus.svelte";
  import {
    type AddDataConfig,
    AddDataStep,
    type AddDataState,
    type ImportStepConfig,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { transitionToNextStep } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ConnectorHeader from "@rilldata/web-common/features/add-data/manager/ConnectorHeader.svelte";
  import ImportDataStatus from "@rilldata/web-common/features/add-data/manager/ImportDataStatus.svelte";
  import SwitchConnectorConfirmation from "@rilldata/web-common/features/add-data/manager/SwitchConnectorConfirmation.svelte";

  export let config: AddDataConfig = {};
  export let initStepState: AddDataState = {
    step: AddDataStep.SelectConnector,
  };
  export let onClose: () => void = () => {};
  export let onStepChange: (step: AddDataStep) => void = () => {};

  const runtimeClient = useRuntimeClient();

  let stepState: AddDataState = initStepState;
  let stateStack: AddDataState[] = [];
  function pushState(newState: AddDataState) {
    stateStack.push(stepState);
    stepState = newState;
  }
  function popState() {
    stepState = stateStack.pop() ?? { step: AddDataStep.SelectConnector };
  }

  // beforeNavigate, onNavigate or afterNavigate do not seem to get called when state changes.
  // "popstate" event does not have direct access to the page state, rather it is under `sveltekit:states` key which seems like internal key.
  // So we need this reactive statement to update the state.
  $: if ("step" in $page.state) {
    stepState = $page.state as AddDataState;
    onStepChange(stepState.step);
  }

  $: schema = (stepState as any).schema as string | undefined;
  $: connector = (stepState as any).connector as string | undefined;

  const SizeClassMap: Partial<Record<AddDataStep, string>> = {
    [AddDataStep.SelectConnector]: "h-fit w-[900px]",
    [AddDataStep.Import]: "h-fit w-[550px]",
  };
  $: sizeClass = SizeClassMap[stepState.step] ?? "h-[630px] w-[900px]";
  $: shouldShowHeader =
    stepState.step === AddDataStep.CreateConnector ||
    stepState.step === AddDataStep.CreateModel ||
    stepState.step === AddDataStep.ExploreConnector;

  let showSwitchConnectorConfirmation = false;

  async function transitionFromInit(
    schema: string | undefined,
    connector: string | undefined,
  ) {
    const newState = await transitionToNextStep(
      runtimeClient,
      {
        step: AddDataStep.SelectConnector,
      },
      {
        schema,
        connector,
      },
    );
    pushState(newState);
  }

  async function transitionToSchema(schema: string) {
    const newState = await transitionToNextStep(runtimeClient, stepState, {
      schema,
    });
    pushState(newState);
  }

  async function setAndStartImport(importConfig: ImportStepConfig) {
    const newState = await transitionToNextStep(runtimeClient, stepState, {
      importConfig,
    });
    pushState(newState);
  }

  function showSwitchConnector() {
    showSwitchConnectorConfirmation = true;
  }

  async function transitionToSelectConnector() {
    const newState = await transitionToNextStep(
      runtimeClient,
      {
        step: AddDataStep.SelectConnector,
      },
      {},
    );
    pushState(newState);
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
      onConnectorChange={(newConnector) =>
        transitionFromInit(undefined, newConnector)}
      onNewConnector={() => transitionFromInit(schema, undefined)}
    />
  {/if}

  {#if stepState.step === AddDataStep.SelectConnector}
    <SourceSelector {config} onSelect={transitionToSchema} onBack={popState} />
  {:else if stepState.step === AddDataStep.CreateConnector}
    <ConnectorForm
      step={stepState}
      onSubmit={pushState}
      onBack={popState}
      {onClose}
    />
  {:else if stepState.step === AddDataStep.CreateModel}
    {#key stepState.connector}
      <SourceForm
        {config}
        step={stepState}
        onSubmit={setAndStartImport}
        onBack={showSwitchConnector}
      />
    {/key}
  {:else if stepState.step === AddDataStep.ExploreConnector}
    {#key stepState.connector}
      <ImportTableForm
        {config}
        step={stepState}
        onSubmit={setAndStartImport}
        onBack={showSwitchConnector}
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
        onBack={popState}
        {onClose}
      />
    {/if}
  {/if}
</div>

{#if stepState.step === AddDataStep.CreateModel || stepState.step === AddDataStep.ExploreConnector}
  <SwitchConnectorConfirmation
    bind:open={showSwitchConnectorConfirmation}
    {stepState}
    onClose={() => void transitionToSelectConnector()}
  />
{/if}
