<script lang="ts">
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";

  export let alertName: string;
</script>

<button
  class="text-fg-secondary hover:text-primary-500 hover:bg-surface-hover grid place-items-center rounded-sm"
  on:click={() => console.log(`open actions for ${alertName}`)}
>
  <ThreeDot size="16px" />
</button>

<!-- TODO: Menu for admins -->

<!-- TODO: Menu for viewers -->
