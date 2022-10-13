<script>
  import { page } from "$app/stores";
</script>

<div class="grid h-screen place-content-center bg-white text-xl">
  {#if $page.error.message}
    <p class="error">{$page.status}: {$page.error.message}</p>
  {:else}
    <p class="error">Encountered a {$page.status} error</p>
  {/if}
</div>
