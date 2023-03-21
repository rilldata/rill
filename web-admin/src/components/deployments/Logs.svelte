<script lang="ts">
  export let logs: string;

  interface Error {
    code: string;
    message: string;
    filePath: string;
  }

  function parseLogs(logs: string): Error[] {
    try {
      return JSON.parse(logs).errors;
    } catch (e) {
      return [];
    }
  }

  $: errors = parseLogs(logs);
</script>

<h2 class="text-lg font-medium">Logs</h2>
{#if logs && errors.length > 0}
  <ul>
    {#each errors as error}
      <li class="border border-red-500 p-2 rounded my-2">
        <p><strong>Code:</strong> {error.code}</p>
        <p><strong>Message:</strong> {error.message}</p>
        <p><strong>File Path:</strong> {error.filePath}</p>
      </li>
    {/each}
  </ul>
{:else}
  <p class="text-gray-500">No logs available.</p>
{/if}
