<!-- web-admin/src/features/admin/shared/SearchInput.svelte -->
<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let placeholder: string = "Search...";
  export let value: string = "";
  export let debounceMs: number = 300;

  const dispatch = createEventDispatcher<{ search: string }>();

  let timeout: ReturnType<typeof setTimeout>;

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    value = target.value;
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

<div class="search-container">
  <input
    type="text"
    class="search-input"
    {placeholder}
    {value}
    on:input={handleInput}
    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
  />
</div>

<style lang="postcss">
  .search-container {
    @apply relative;
  }

  .search-input {
    @apply w-full px-3 py-2 text-sm rounded-md border border-slate-300
      dark:border-slate-600 bg-white dark:bg-slate-800
      text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 dark:placeholder:text-slate-500
      focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }
</style>
