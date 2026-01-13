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
  import {
    ResourceKind,
    type UserFacingResourceKinds,
  } from "../entity-management/resource-selectors";
  import { builderActions } from "bits-ui";
  import { GitBranch } from "lucide-svelte";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";

  const downstreamMapping = new Map([
    [ResourceKind.MetricsView, new Set([ResourceKind.Explore])],
    [ResourceKind.Source, new Set([ResourceKind.Model])],
    [
      ResourceKind.Model,
      new Set([ResourceKind.MetricsView, ResourceKind.Model]),
    ],
  ]);

  export let resources: (V1Resource | undefined)[];
  export let allResources: V1Resource[];
  export let selectedResource: V1Resource | undefined = undefined;
  export let current = false;
  export let downstream = false;
  export let upstream = false;
  export let filePath: string = "";
  export let graphSupported = false;
  export let openGraph: (() => void) | null = null;

  let open = false;

  $: dropdown = resources.length > 1;

  $: exampleResource = selectedResource ?? resources?.[0];

  $: meta = exampleResource?.meta;

  $: resourceKind = meta?.name?.kind as ResourceKind | undefined;
  $: resourceName = meta?.name?.name ?? filePath?.split("/").pop();

  $: withoutComponents = resources?.filter((r) => !r?.component);

  $: componentsOnly = !withoutComponents.length && resources.length;

  $: allRefs = resources?.map((r) => r?.meta?.refs).flat();

  $: upstreamResources = downstream
    ? []
    : allResources.filter(({ meta }) =>
        allRefs?.find(
          (ref) =>
            ref?.name === meta?.name?.name && ref?.kind === meta?.name?.kind,
        ),
      );

  $: downstreamKinds =
    !upstream && resourceKind && downstreamMapping.get(resourceKind);

  $: downstreamResources = downstreamKinds
    ? allResources.filter(({ meta }) => {
        const kind = meta?.name?.kind as UserFacingResourceKinds | undefined;
        if (!kind) return false;
        return (
          downstreamKinds.has(kind) &&
          meta?.refs?.find(({ kind, name }) => {
            return selectedResource
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

  function generateLabel(resources: (V1Resource | undefined)[]) {
    const counts: Map<UserFacingResourceKinds, number> = new Map();

    for (const r of resources) {
      const kind = r?.meta?.name?.kind as UserFacingResourceKinds | undefined;
      if (!kind) continue;
      counts.set(kind, (counts.get(kind) ?? 0) + 1);
    }

    return Array.from(counts)
      .map(
        ([kind, count]) =>
          `${count} ${ResourceKindMap[kind].folderName.slice(0, -1) + (count > 1 ? "s" : "")}`,
      )
      .join(", ");
  }

</script>

{#if upstreamResources.length}
  <svelte:self resources={upstreamResources} {allResources} upstream />

  {#if !componentsOnly}
    <CaretDownIcon size="12px" className="text-gray-500 -rotate-90 flex-none" />
  {/if}
{/if}

{#if !componentsOnly}
  <div class="crumb">
    <div class="crumb__trigger">
      <DropdownMenu.Root bind:open>
        <DropdownMenu.Trigger asChild let:builder>
          <svelte:element
            this={dropdown ? "button" : "a"}
            class:open
            class="text-gray-500 px-[5px] py-1 w-full max-w-fit line-clamp-1"
            class:selected={current}
            href={dropdown
              ? undefined
              : exampleResource
                ? $previewModeStore
                  ? `/edit?resource=${resourceName}`
                  : `/files${exampleResource?.meta?.filePaths?.[0] ?? filePath}`
                : "#"}
            {...dropdown ? builder : {}}
            use:builderActions={{ builders: dropdown ? [builder] : [] }}
          >
            <CrumbTrigger
              {filePath}
              kind={resourceKind}
              label={!selectedResource && dropdown
                ? generateLabel(resources)
                : resourceName}
            />
          </svelte:element>
        </DropdownMenu.Trigger>

        {#if dropdown}
          <DropdownMenu.Content align="start">
            {#each resources as resource (resource?.meta?.name?.name)}
              {@const kind = resource?.meta?.name?.kind}
              <DropdownMenu.Item
                href={$previewModeStore
                  ? `/edit?resource=${resource?.meta?.name?.name}`
                  : `/files${resource?.meta?.filePaths?.[0] ?? filePath}`}
              >
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
    </div>
    {#if current && graphSupported && openGraph}
      <button
        type="button"
        class="graph-trigger"
        on:click={openGraph}
        aria-label="Open resource graph"
      >
        <GitBranch size="13px" aria-hidden="true" />
        <span class="sr-only">Open resource graph</span>
      </button>
    {/if}
  </div>
{/if}

{#if downstreamResources.length}
  <CaretDownIcon size="12px" className="text-gray-500 -rotate-90 flex-none" />

  <svelte:self resources={downstreamResources} {allResources} downstream />
{/if}

<style lang="postcss">
  .crumb {
    @apply inline-flex items-center gap-x-1 min-w-0;
  }

  .crumb__trigger {
    @apply flex-1 min-w-0;
  }

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

  .graph-trigger {
    @apply flex-none inline-flex items-center justify-center rounded-md border transition-colors shadow-sm ml-1 px-2 py-[3px];
    border-color: var(--border, #e5e7eb);
    background-color: var(--surface, #ffffff);
    color: var(--muted-foreground, #6b7280);
    min-width: 30px;
    height: 26px;
  }

  .graph-trigger:hover {
    color: var(--foreground, #1f2937);
    border-color: color-mix(
      in srgb,
      var(--border, #e5e7eb) 70%,
      var(--foreground, #1f2937)
    );
  }

  .graph-trigger:focus-visible {
    @apply outline-none ring ring-offset-1;
    ring-color: var(--ring, #93c5fd);
    ring-offset-color: var(--surface, #ffffff);
  }
</style>
