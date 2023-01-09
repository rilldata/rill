<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { slide } from "svelte/transition";
  import EmbeddedSourceLink from "./EmbeddedSourceLink.svelte";
  import SourceDomain from "./SourceDomain.svelte";

  export let sources;
  export let location: string;
  export let type: string;
  export let active = true;
</script>

<div>
  <SourceDomain bind:active {type} {location} {sources} />
  {#if active}
    <div transition:slide={{ duration: LIST_SLIDE_DURATION }} class="pb-2">
      {#each sources as { uri, abbreviatedURI }}
        <EmbeddedSourceLink {uri} {abbreviatedURI} />
      {/each}
    </div>
  {/if}
</div>
