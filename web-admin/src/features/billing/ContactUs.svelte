<script lang="ts">
  export let variant: "normal" | "enterprise" = "normal";

  $: text = variant === "normal" ? "contact us" : "contact support";

  function onClick() {
    window.Pylon("show");
  }
</script>

<button on:click={onClick} class="text-primary-600 text-sm ml-0.5">
  {text}
</button>
