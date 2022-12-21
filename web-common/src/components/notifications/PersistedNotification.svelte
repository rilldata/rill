<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { scale } from "svelte/transition";
  import { IconButton } from "../button";
  import Close from "../icons/Close.svelte";

  const dispatch = createEventDispatcher();
</script>

<div
  transition:scale={{ duration: 200, start: 0.98, opacity: 0 }}
  class="fixed bottom-10 left-1/2 -translate-x-1/2 max-w-4xl py-5 text-sm whitespace-normal
         text-slate-100 bg-slate-800 shadow-lg rounded-md flex flex-col gap-5"
>
  <div class="flex flex-row items-center justify-between pl-6 pr-4">
    <h1><slot name="title" /></h1>
    <div>
      <IconButton
        marginClasses="ml-3"
        on:click={() => dispatch("clear")}
        bgDark
      >
        <Close size="16px" color="#fff" />
      </IconButton>
    </div>
  </div>
  <hr class="border-slate-200" />
  <div class="font-normal px-7 py-3">
    <slot name="body" />
  </div>
</div>
