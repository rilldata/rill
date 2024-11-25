<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { type V1Resource } from "@rilldata/web-common/runtime-client";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../entity-management/resource-icon-mapping";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { ResourceKindMap } from "../file-explorer/new-files";
  import CrumbTrigger from "./CrumbTrigger.svelte";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { builderActions } from "bits-ui";

  const downstreamMapping = new Map([
    [ResourceKind.MetricsView, ResourceKind.Explore],
    [ResourceKind.Source, ResourceKind.Model],
    [ResourceKind.Model, ResourceKind.MetricsView],
  ]);

  export let resources: (V1Resource | undefined)[];
  export let allResources: V1Resource[];
  export let selected = false;
  export let downstream = false;
  export let upstream = false;
  export let filePath: string = "";

  let open = false;

  $: firstResource = resources?.[0];

  $: resourceKind = firstResource?.meta?.name?.kind as ResourceKind | undefined;
  $: resourceName =
    firstResource?.meta?.name?.name ?? filePath?.split("/").pop();

  $: allRefs = resources?.map((r) => r?.meta?.refs).flat();

  $: referencedResources = downstream
    ? []
    : allResources.filter(({ meta }) =>
        allRefs?.find(
          (ref) =>
            ref?.name === meta?.name?.name && ref?.kind === meta?.name?.kind,
        ),
      );

  $: dropdown = resources.length > 1;

  $: downstreamKind = resourceKind && downstreamMapping.get(resourceKind);

  $: downstreamResources = downstreamKind
    ? allResources.filter(({ meta }) => {
        return (
          meta?.name?.kind === downstreamKind &&
          meta?.refs?.find(({ kind, name }) => {
            return selected
              ? kind === resourceKind && name === resourceName
              : resources.find(
                  (r) =>
                    r?.meta?.name?.kind === kind &&
                    r?.meta?.name?.name === name,
                );
          })
        );
      })
    : [];
</script>

{#if !downstream && referencedResources.length}
  <svelte:self resources={referencedResources} {allResources} upstream />
{/if}

{#if !downstream && referencedResources.length}
  <CaretDownIcon size="12px" className="text-gray-500 -rotate-90 flex-none" />
{/if}

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <svelte:element
      this={dropdown ? "button" : "a"}
      class:open
      class="text-gray-500 px-1.5 py-1 w-full max-w-fit line-clamp-1"
      class:selected
      href={dropdown
        ? undefined
        : firstResource
          ? `/files${firstResource?.meta?.filePaths?.[0]}`
          : "#"}
      {...dropdown ? builder : {}}
      use:builderActions={{ builders: dropdown ? [builder] : [] }}
    >
      <CrumbTrigger
        {filePath}
        kind={resourceKind}
        label={!selected && dropdown
          ? `${resources?.length} ${ResourceKindMap[resourceKind ?? ResourceKind.Component].folderName}`
          : resourceName}
      />
    </svelte:element>
  </DropdownMenu.Trigger>

  {#if dropdown}
    <DropdownMenu.Content align="start">
      {#each resources as resource (resource?.meta?.name?.name)}
        {@const kind = resource?.meta?.name?.kind}
        <DropdownMenu.Item href="/files{resource?.meta?.filePaths?.[0] ?? '/'}">
          {#if kind}
            <svelte:component
              this={resourceIconMapping[kind]}
              color={resourceColorMapping[kind]}
              size="14px"
            />
          {/if}
          {resource?.meta?.name?.name}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  {/if}
</DropdownMenu.Root>

{#if !upstream && downstreamResources.length}
  <CaretDownIcon size="12px" className="text-gray-500 -rotate-90 flex-none" />
{/if}

{#if !upstream && downstreamResources.length}
  <svelte:self resources={downstreamResources} {allResources} downstream />
{/if}

<style lang="postcss">
  a:hover,
  button:hover {
    @apply text-gray-700;
  }

  .selected {
    @apply text-gray-900;
  }

  .open {
    @apply bg-slate-200 rounded-[2px] text-gray-700;
  }
</style>
