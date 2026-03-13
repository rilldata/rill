<script lang="ts">
  export let name: string;
  export let onClick: () => void;
</script>

<button
  class="text-primary-600 hover:text-primary-800 hover:underline text-left"
  on:click={onClick}
>
  {name}
</button>
