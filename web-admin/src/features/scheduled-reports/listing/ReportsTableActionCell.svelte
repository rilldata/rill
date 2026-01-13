<script lang="ts">
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";

  export let reportName: string;
</script>

<button
  class="text-muted-foreground hover:text-primary-500 hover:bg-gray-200 grid place-items-center rounded-sm"
  on:click={() => console.log(`open actions for ${reportName}`)}
>
  <ThreeDot size="16px" />
</button>

<!-- TODO: Menu for admins -->

<!-- TODO: Menu for viewers -->
