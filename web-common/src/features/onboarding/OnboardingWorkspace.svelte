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
  import { PresentationIcon } from "lucide-svelte";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";

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

<div class="flex flex-col pt-20 mx-auto px-8 gap-y-6 items-center h-full w-fit">
  <div class="flex flex-row text-center gap-x-8">
    <div
      class="flex flex-col w-64 p-6 gap-y-4 bg-card border rounded-md shadow-sm"
    >
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

    <div class="flex flex-col w-64 gap-y-8">
      <GenerateSampleData />
      <Button
        type="secondary"
        onClick={() => createResourceAndNavigate(ResourceKind.Model)}
        large
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.Model]}
          color={resourceColorMapping[ResourceKind.Model]}
          size="16px"
        />
        Create blank model
      </Button>
      <Button
        type="secondary"
        onClick={() => createResourceAndNavigate(ResourceKind.MetricsView)}
        large
      >
        <svelte:component
          this={resourceIconMapping[ResourceKind.MetricsView]}
          color={resourceColorMapping[ResourceKind.MetricsView]}
          size="16px"
        />
        Create a metrics view
      </Button>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <Button type="secondary" large builders={[builder]}>
            <PresentationIcon size="16px" />
            Try demo projects
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content side="right" align="start">
          {#each EXAMPLES as example (example.name)}
            <DropdownMenu.Item on:click={() => unpackProject(example)}>
              {example.name}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  </div>

  <div class="h-px w-full border border-muted"></div>

  <div class="flex flex-col gap-y-2 text-xs text-slate-500">
    <div class="font-semibold w-full text-center">
      Tips for data workflow in rill
    </div>
    <ul class="list-decimal">
      <li>Import data – Add or drag files (Parquet, NDJSON, CSV).</li>
      <li>Model sources – Combine and shape data with SQL.</li>
      <li>Define metrics – Create metrics and dimensions.</li>
      <li>Explore insights – Visualize data in interactive dashboards.</li>
    </ul>
  </div>
</div>
