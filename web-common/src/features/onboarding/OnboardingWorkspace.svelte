<script lang="ts">
  import { goto } from "$app/navigation";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Button from "../../components/button/Button.svelte";
  import { createRuntimeServiceUnpackExample } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { addSourceModal } from "../sources/modal/add-source-visibility";
  import ImportData from "@rilldata/web-common/components/icons/ImportData.svelte";
  import GenerateSampleData from "@rilldata/web-common/features/sample-data/GenerateSampleData.svelte";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createResourceAndNavigate } from "@rilldata/web-common/features/file-explorer/new-files.ts";
  import { EXAMPLES } from "@rilldata/web-common/features/welcome/constants.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
  import { LightbulbIcon, PresentationIcon } from "lucide-svelte";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { builderActions, getAttrs } from "bits-ui";

  $: ({ instanceId } = $runtime);

  const unpackExampleProject = createRuntimeServiceUnpackExample();

  async function unpackProject(example: (typeof EXAMPLES)[number]) {
    await behaviourEvent?.fireSplashEvent(
      example
        ? BehaviourEventAction.ExampleAdd
        : BehaviourEventAction.ProjectEmpty,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      example?.name,
    );

    try {
      await $unpackExampleProject.mutateAsync({
        instanceId,
        data: {
          name: example.name,
          force: true,
        },
      });

      await waitUntil(() => fileArtifacts.hasFileArtifact(example.firstFile));
      await goto(`/files${example.firstFile}`);
    } catch {
      // no-op
    }
  }
</script>

<div class="container">
  <div class="cta-container">
    <div class="import-data-container cta-item">
      <div class="flex flex-col gap-y-1">
        <div class="font-semibold text-base">Import data</div>
        <div class="text-xs">
          Add or drag a file here (Parquet, NDJSON, CSV).
        </div>
      </div>
      <div class="mx-auto">
        <ImportData />
      </div>
      <Button type="primary" onClick={addSourceModal.open}>+ Add Data</Button>
    </div>

    <div class="my-auto text-gray-400 text-base">or</div>

    <div class="flex flex-col w-64 gap-y-4">
      <GenerateSampleData type="home" />
      <button
        class="cta-button cta-item"
        on:click={() => createResourceAndNavigate(ResourceKind.Model)}
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.Model]}
          color={resourceColorMapping[ResourceKind.Model]}
          size="16px"
        />
        Create blank model
      </button>
      <button
        class="cta-button cta-item"
        on:click={() => createResourceAndNavigate(ResourceKind.MetricsView)}
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.MetricsView]}
          color={resourceColorMapping[ResourceKind.MetricsView]}
          size="16px"
        />
        Create a metrics view
      </button>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button
            class="cta-button cta-item"
            {...getAttrs([builder])}
            use:builderActions={{ builders: [builder] }}
          >
            <PresentationIcon size="16px" />
            Try demo projects
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content side="right" align="start">
          {#each EXAMPLES as example (example.name)}
            <DropdownMenu.Item on:click={() => unpackProject(example)}>
              {example.title}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  </div>

  <div class="flex flex-row gap-x-8 items-center w-full">
    <div class="h-px grow border border-muted"></div>
    <LightbulbIcon class="text-border" size="16px" />
    <div class="h-px grow border border-muted"></div>
  </div>

  <div class="flex flex-col mx-auto w-fit gap-y-2 text-xs text-slate-500">
    <div class="font-semibold text-center">Tips for data workflow in rill</div>
    <ul class="list-decimal">
      <li>Import data – Add or drag files (Parquet, NDJSON, CSV).</li>
      <li>Model sources – Combine and shape data with SQL.</li>
      <li>Define metrics – Create metrics and dimensions.</li>
      <li>Explore insights – Visualize data in interactive dashboards.</li>
    </ul>
  </div>
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col m-auto px-8 gap-y-6 w-fit;
  }

  .cta-container {
    @apply flex flex-row text-center gap-x-6;
  }

  .import-data-container {
    @apply flex flex-col w-64 p-6 gap-y-4;
  }

  .cta-item {
    @apply bg-card border rounded-md shadow-sm;
  }

  .cta-button {
    @apply flex flex-row text-center items-center justify-center;
    @apply text-sm gap-x-2 h-12;
  }
</style>
