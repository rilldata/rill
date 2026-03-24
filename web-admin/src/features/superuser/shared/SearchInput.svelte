<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let placeholder: string = "Search...";
  export let value: string = "";
  export let debounceMs: number = 300;

  const dispatch = createEventDispatcher<{ search: string }>();

  let timeout: ReturnType<typeof setTimeout>;

  function handleInput() {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      dispatch("search", value);
    }, debounceMs);
  }

  function handleSubmit() {
    clearTimeout(timeout);
    dispatch("search", value);
  }
</script>

<div class="relative">
  <input
    type="text"
    class="w-full px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
    {placeholder}
    bind:value
    on:input={handleInput}
    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
  />
</div>
