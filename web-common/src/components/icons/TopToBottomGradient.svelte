<script lang="ts">
  export let startColor = "currentColor";
  export let endColor;
</script>

{#if endColor}
  <defs>
    <linearGradient id="top-bottom-grad" x1="100%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color={startColor} />
      <stop offset="100%" style="stop-color:var(--color-{endColor});" />
    </linearGradient>
  </defs>
{/if}
