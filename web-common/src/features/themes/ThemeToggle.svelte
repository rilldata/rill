<script lang="ts">
  import { ComputerIcon, MoonIcon, SunIcon } from "lucide-svelte";

  import { viewModeStore } from "./actions";

  $: viewMode = $viewModeStore;
</script>

<div class="container">
  <!-- <button
        title="Match system theme"
        class:selected={viewMode === "system"}
        on:click={() => {
          viewModeStore.set("system");
          const prefersDark = window.matchMedia("(prefers-color-scheme: dark)");
    
          if (prefersDark.matches) {
            document.documentElement.setAttribute("data-theme", "dark");
          } else {
            document.documentElement.setAttribute("data-theme", "light");
          }
        }}
      >
        <ComputerIcon size="12px" />
      </button> -->

  <button
    class:selected={viewMode === "light"}
    on:click={() => {
      viewModeStore.set("light");
    }}
  >
    <SunIcon size="12px" />
  </button>
  <button
    class:selected={viewMode === "dark"}
    on:click={() => {
      viewModeStore.set("dark");
    }}
  >
    <MoonIcon size="12px" />
  </button>
</div>

<style lang="postcss">
  .container {
    @apply flex border border-gray-200 rounded-lg h-6 w-fit items-center flex-none;
    @apply bg-background;
  }

  button {
    @apply size-[24px] grid place-content-center overflow-hidden;
    @apply text-gray-500 bg-transparent;
  }

  button:hover {
    @apply text-gray-900;
  }

  button.selected {
    @apply outline-gray-300 bg-surface outline outline-1 rounded-lg;
    @apply text-gray-900;
  }
</style>
