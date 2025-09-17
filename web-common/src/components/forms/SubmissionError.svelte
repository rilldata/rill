<script lang="ts">
  import { AlertCircleIcon } from "lucide-svelte";

  export let message: string;
  export let details: string | undefined = undefined;

  let showDetails = true;

  function toggleDetails() {
    showDetails = !showDetails;
  }
</script>

<div class="error-container">
  <div class="flex items-start gap-1 min-w-0">
    <div class="flex-shrink-0 flex items-start">
      <AlertCircleIcon size={22} class="text-red-600" />
    </div>
    <div class="flex-1 min-w-0">
      <div class="message text-gray-700 font-normal text-sm">
        {message}
      </div>
      {#if details}
        <button
          class="flex items-center mt-2 cursor-pointer select-none"
          on:click={toggleDetails}
        >
          <span class="text-xs font-semibold text-gray-500 capitalize"
            >Connection error</span
          >
          <div class="icon-wrapper ml-1">
            <svg
              width="10"
              height="10"
              viewBox="0 0 10 10"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              style="transform: rotate({showDetails
                ? 180
                : 0}deg); transition: transform 0.2s;"
            >
              <g clip-path="url(#clip0_1706_319718)">
                <rect width="10" height="10" fill="white" fill-opacity="0.01" />
                <path
                  d="M8.13793 3.20105C8.31782 3.20142 8.41243 3.41467 8.29223 3.54871L4.99828 7.22156L1.71313 3.54871C1.59309 3.41449 1.68834 3.20105 1.8684 3.20105H8.13793Z"
                  fill="#6B7280"
                />
              </g>
              <defs>
                <clipPath id="clip0_1706_319718">
                  <rect width="10" height="10" fill="white" />
                </clipPath>
              </defs>
            </svg>
          </div>
        </button>
        {#if showDetails}
          <div class="details-section border-l-2 border-gray-300">
            <pre class="details whitespace-pre-wrap break-words">{details}</pre>
          </div>
        {/if}
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .error-container {
    @apply border-red-600 bg-red-50;
    @apply p-2 flex border rounded;
    @apply max-h-48 overflow-y-auto;
  }

  .message {
    @apply whitespace-pre-wrap break-words;
  }

  .details-section {
    @apply mt-2 text-xs;
    @apply pl-2;
  }

  .details {
    @apply text-gray-600 whitespace-pre-wrap break-words;
  }
</style>
