<!--
  A text component with a continuous left-to-right shimmer wave animation.
  Used for loading states like "Thinking..." and "Planning next moves...".
-->
<script lang="ts">
  export let className = "";
</script>

<span class="shimmer-text {className}">
  <slot />
</span>

<style lang="postcss">
  .shimmer-text {
    background: linear-gradient(
      90deg,
      currentColor 0%,
      currentColor 40%,
      rgba(107, 114, 128, 0.4) 50%,
      currentColor 60%,
      currentColor 100%
    );
    background-size: 200% 100%;
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    animation: shimmer-wave 5s linear infinite;
  }

  @keyframes shimmer-wave {
    from {
      background-position: 200% 0;
    }
    to {
      background-position: -200% 0;
    }
  }
</style>
