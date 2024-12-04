<script lang="ts">
  function onClick() {
    window.Pylon("show");
  }
</script>

<p class="flex flex-row text-sm text-gray-500 text-center">
  <span>Need help?</span>
  <button on:click={onClick} class="text-primary-600 text-sm ml-0.5">
    Contact us
  </button>
</p>
