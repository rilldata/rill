<!--
  Animated indicator shown while the AI is still working.
  Uses animated dots to avoid misrepresenting what's happening technically.
-->
<div class="working-block">
  <div class="dots">
    <span class="dot"></span>
    <span class="dot"></span>
    <span class="dot"></span>
  </div>
</div>

<style lang="postcss">
  .working-block {
    @apply w-full max-w-full self-start py-2;
  }

  .dots {
    @apply flex items-center gap-1;
  }

  .dot {
    @apply w-1.5 h-1.5 rounded-full bg-gray-400;
    animation: pulse 1.4s ease-in-out infinite;
  }

  .dot:nth-child(1) {
    animation-delay: 0s;
  }

  .dot:nth-child(2) {
    animation-delay: 0.2s;
  }

  .dot:nth-child(3) {
    animation-delay: 0.4s;
  }

  @keyframes pulse {
    0%,
    80%,
    100% {
      opacity: 0.3;
      transform: scale(0.8);
    }
    40% {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
