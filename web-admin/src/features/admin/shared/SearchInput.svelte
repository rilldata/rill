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

<div class="relative">
  <input
    type="text"
    class="w-full px-3 py-2 text-sm rounded-md border border-slate-300 bg-slate-50 text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    {placeholder}
    {value}
    on:input={handleInput}
    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
  />
</div>
