<script lang="ts">
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { slide } from "svelte/transition";
  import type { SourceURI } from "../../group-uris";
  import EmbeddedSourceLink from "./EmbeddedSourceLink.svelte";
  import SourceDomain from "./SourceDomain.svelte";

  export let sources: SourceURI[];
  export let location: string;
  export let connector: string;
  export let active = true;
</script>

<div>
  <SourceDomain bind:active {connector} {location} {sources} />
  {#if active}
    <div
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
      class="pb-2"
    >
      {#each sources as { uri, abbreviatedURI, name }}
        <EmbeddedSourceLink
          {uri}
          {abbreviatedURI}
          {connector}
          cachedSourceName={name}
        />
      {/each}
    </div>
  {/if}
</div>
