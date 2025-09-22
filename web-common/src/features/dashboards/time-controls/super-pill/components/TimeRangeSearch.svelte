<script lang="ts">
  import SyntaxElement from "./SyntaxElement.svelte";
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import { Clock } from "lucide-svelte";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import { ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";

  const message = "Unable to parse time string";

  export let context: string;
  export let width: number;
  export let inError: boolean;
  export let timeString: string | undefined = undefined;
  export let searchValue = timeString || "";
  export let onSelectRange: (range: string) => void;

  let searchElement: HTMLInputElement;
  let unableToParse = false;

  const latestNSearches = localStorageStore(`${context}-recent-searches`, [
    "-7d/d to -3d/d",
    "-1d/d to -1d/d+6h",
    "D3 of M11",
  ]);

  export function updateSearch(value: string) {
    searchValue = value;
    searchElement.focus();
  }
</script>

<div
  class="border-b h-fit pt-2.5 py-0 flex p-3 gap-y-2 flex-col overflow-y-auto"
  style:width="{width}px"
>
  <form
    class:error={(inError || unableToParse) &&
      timeString !== ALL_TIME_RANGE_ALIAS}
    class=""
    on:submit={(e) => {
      e.preventDefault();

      if (searchValue === ALL_TIME_RANGE_ALIAS) {
        onSelectRange(ALL_TIME_RANGE_ALIAS);
        searchValue = "";
        unableToParse = false;
        return;
      }

      try {
        parseRillTime(searchValue);

        unableToParse = false;

        latestNSearches.update((searches) => {
          return Array.from(new Set([searchValue, ...searches].slice(0, 15)));
        });

        onSelectRange(searchValue);

        searchValue = "";
      } catch (e) {
        console.error(e);
        unableToParse = true;
      }
    }}
  >
    <span
      class="mr-1 flex-none"
      role="presentation"
      on:click={() => {
        searchElement.focus();
      }}
    >
      <Clock size={15} />
    </span>
    <input
      placeholder="Enter a time range"
      type="text"
      class="h-7 border w-full"
      on:keydown={() => {
        if (unableToParse) {
          unableToParse = false;
        }
      }}
      bind:this={searchElement}
      bind:value={searchValue}
    />
  </form>

  {#if unableToParse}
    <div class="text-red-500 text-xs">{message}</div>
  {/if}

  <div class="flex gap-x-2 size-full overflow-x-auto pb-2.5">
    {#each $latestNSearches as search, i (i)}
      <SyntaxElement range={search} onClick={updateSearch} />
    {/each}
  </div>
</div>

<style lang="postcss">
  form {
    @apply overflow-hidden;
    @apply flex justify-center gap-x-1 items-center pl-2 pr-0.5;
    @apply bg-background justify-center;
    @apply border border-gray-300 rounded-sm;
    @apply h-7 w-full truncate;
  }

  form:focus-within {
    @apply border-primary-500;
  }

  form.error {
    @apply border-red-500;
  }

  input {
    @apply p-0 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }
</style>
