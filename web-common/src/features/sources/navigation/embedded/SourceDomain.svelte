<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import { fly } from "svelte/transition";
  import type { SourceURI } from "../../group-uris";
  import SourceTypeLabel from "./SourceTypeLabel.svelte";
  export let active = true;
  export let connector: string;
  export let location: string;
  export let sources: SourceURI[];
  export let expandable = true;
</script>

<div class="w-full">
  <CollapsibleSectionTitle
    bind:active
    suppressTooltip={!expandable}
    tooltipText="sources"
  >
    <div
      class="grid items-center gap-x-2 justify-start justify-items-start pl-3 pr-3 font-normal"
      style:grid-template-columns="auto 1fr max-content"
      style:height="24px"
    >
      <div>
        <SourceTypeLabel {connector} />
      </div>
      <div
        class="text-left w-full grow text-ellipsis overflow-hidden whitespace-nowrap text-gray-600"
      >
        {location}
      </div>
      <div style:min-width="16px" class="text-right">
        {#if !active}<div
            transition:fly|local={{ duration: LIST_SLIDE_DURATION, y: 4 }}
          >
            {sources.length}
          </div>{/if}
      </div>
    </div>
  </CollapsibleSectionTitle>
</div>
