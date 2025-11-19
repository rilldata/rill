<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { slideRight } from "../../../lib/transitions";

  export let value: string;
  export let onSubmit: () => void;
  export let onClose: () => void;

  let isSearchElementOpen = false;

  function _onClose() {
    isSearchElementOpen = false;
    onClose();
  }
</script>

{#if !isSearchElementOpen}
  <Button
    type="toolbar"
    onClick={() => (isSearchElementOpen = !isSearchElementOpen)}
  >
    <SearchIcon size="16px" />
    <span>Search</span>
  </Button>
{:else}
  <div transition:slideRight={{}} class="flex items-center gap-x-1">
    <Search bind:value {onSubmit} />
    <button
      on:click={_onClose}
      class="p-1.5 rounded hover:bg-gray-100 transition-colors"
    >
      <Close className="ui-copy-icon" />
    </button>
  </div>
{/if}
