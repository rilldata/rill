<script>
  import { page } from "$app/stores";
</script>

<div class="w-full h-full items-center flex justify-center">
  {#if $page.error?.message}
    <p class="error">{$page.status}: {$page.error?.message}</p>
  {:else}
    <p class="error">Encountered a {$page.status} error</p>
  {/if}
</div>
