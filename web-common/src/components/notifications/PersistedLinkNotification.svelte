<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { scale } from "svelte/transition";
  import { IconButton } from "../button";
  import Check from "../icons/Check.svelte";
  import Close from "../icons/Close.svelte";
  import type { Link } from "./notificationStore";

  export let message: string;
  export let link: Link;

  const dispatch = createEventDispatcher();
</script>

<div
  transition:scale={{ duration: 200, start: 0.98, opacity: 0 }}
  class="fixed bottom-10 left-1/2 -translate-x-1/2 py-0.5 bg-gray-800 rounded-sm shadow flex items-center"
>
  <div class="flex items-center px-4 py-1.5 gap-x-1.5">
    <Check size="18px" className="text-white" />
    <span class="text-gray-50 text-sm">
      {message}
    </span>
  </div>
  <div class="px-4 py-1.5 border-l border-gray-600 text-sm">
    <a
      class="text-primary-300 hover:text-primary-200"
      href={link.href}
      on:click={() => dispatch("clear")}>{link.text}</a
    >
  </div>
  <div class="px-2.5 py-1.5 border-l border-gray-600">
    <IconButton on:click={() => dispatch("clear")} bgDark>
      <Close size="18px" color="#fff" />
    </IconButton>
  </div>
</div>
