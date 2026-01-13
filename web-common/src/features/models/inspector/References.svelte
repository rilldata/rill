<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { formatCompactInteger } from "@rilldata/web-common/lib/formatters";
  import {
    type V1ResourceName,
    createQueryServiceTableCardinality,
    createRuntimeServiceGetResource,
  } from "@rilldata/web-common/runtime-client";
  import { derived } from "svelte/store";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { removeLeadingSlash } from "../../entity-management/entity-mappers";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  export let refs: V1ResourceName[];
  export let modelHasError: boolean;

  $: ({ instanceId } = $runtime);

  let showReferences = true;

  $: referencedResourcesStore = derived(
    refs.map((ref) => {
      return createRuntimeServiceGetResource(instanceId, {
        "name.name": ref.name as string,
        "name.kind": ref.kind as string,
      });
    }),
    (refs) => refs.map((ref) => ref.data),
  );
  $: referencedResources = $referencedResourcesStore;

  $: referencedResourceCardinalitiesStore = derived(
    refs.map((ref) => {
      return createQueryServiceTableCardinality(
        instanceId,
        ref.name as string,
        {},
        {
          query: {
            select: (data) => +(data.cardinality ?? 0),
          },
        },
      );
    }),
    (refs) => refs.map((ref) => ref.data),
  );
  $: referencedResourcesCardinalities = $referencedResourceCardinalitiesStore;
</script>

{#if refs.length}
  <div>
    <div class=" pl-4 pr-4">
      <CollapsibleSectionTitle
        tooltipText="References"
        bind:active={showReferences}
      >
        Referenced in this model
      </CollapsibleSectionTitle>
    </div>

    {#if showReferences}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }} class="mt-2">
        {#each refs as reference, index (reference.name)}
          {@const resource = referencedResources[index]}
          {@const cardinality = referencedResourcesCardinalities[index]}
          {@const filePath = resource?.resource?.meta?.filePaths?.[0]}
          {#if filePath}
            <div>
              <WithModelResultTooltip {modelHasError}>
                <a
                  href="/files/{removeLeadingSlash(filePath)}"
                  class="ui-copy-muted grid justify-between gap-x-2 pl-4 pr-4 hover:bg-yellow-200 hover:cursor-pointer"
                  style:grid-template-columns="auto max-content"
                  class:text-muted-foreground={modelHasError}
                >
                  <div class="truncate flex items-center gap-x-2">
                    <div class="truncate">
                      {reference.name}
                    </div>
                  </div>

                  {#if cardinality}
                    <div class="text-muted-foreground">
                      {`${formatCompactInteger(cardinality)} rows`}
                    </div>
                  {/if}
                </a>

                <svelte:fragment slot="tooltip-title">
                  <div class="break-all">
                    {reference.name}
                  </div>
                </svelte:fragment>
                <svelte:fragment slot="tooltip-right">
                  {#if resource?.resource?.source}
                    {resource?.resource?.source?.state?.connector}
                  {/if}
                </svelte:fragment>

                <svelte:fragment slot="tooltip-description">
                  <TooltipShortcutContainer>
                    <div>Open in workspace</div>
                    <Shortcut>Click</Shortcut>
                  </TooltipShortcutContainer>
                </svelte:fragment>
              </WithModelResultTooltip>
            </div>
          {/if}
        {/each}
      </div>
    {/if}
  </div>
{/if}
