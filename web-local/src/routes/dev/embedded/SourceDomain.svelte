<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import { fly } from "svelte/transition";
  import SourceTypeLabel from "./SourceTypeLabel.svelte";
  export let active = true;
  export let type: string;
  export let location: string;
  export let sources: string[];
  export let expandable = true;
</script>

<div class="w-full">
  <CollapsibleSectionTitle
    bind:active
    suppressTooltip={!expandable}
    tooltipText="these sources"
  >
    <div
      class="grid items-center gap-x-2 justify-start justify-items-start pl-2 pr-3 font-normal"
      style:grid-template-columns="auto 1fr max-content"
      style:height="24px"
    >
      <div>
        <SourceTypeLabel {type} />
      </div>
      <div
        class="text-left w-full grow text-ellipsis overflow-hidden whitespace-nowrap text-gray-600 font-medium"
      >
        {location}
      </div>
      <div style:width="24px" class="text-right">
        {#if !active}<div
            transition:fly|local={{ duration: LIST_SLIDE_DURATION, y: 4 }}
          >
            {sources.length}
          </div>{/if}
      </div>
    </div>
  </CollapsibleSectionTitle>
</div>
