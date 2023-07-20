<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { createRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import References from "./References.svelte";

  export let modelName: string;

  $: getModel = createRuntimeServiceGetCatalogEntry(
    $runtime?.instanceId,
    modelName
  );
  let entry;
  // refresh entry value only if the data has changed
  $: entry = $getModel?.data?.entry || entry;

  let showColumns = true;
</script>

<div class="model-profile">
  {#if entry && entry?.model?.sql?.trim()?.length}
    <References {modelName} />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="selected columns"
          bind:active={showColumns}
        >
          Selected columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
          <ColumnProfile objectName={entry?.model?.name} indentLevel={0} />
        </div>
      {/if}
    </div>
  {/if}
</div>
