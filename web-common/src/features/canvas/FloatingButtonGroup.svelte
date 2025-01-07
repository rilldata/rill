<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let show: boolean;
  export let left: number;
  export let top: number;

  const dispatch = createEventDispatcher();
</script>

<div
  class="floating-buttons absolute flex flex-col gap-1 opacity-0 transition-opacity pointer-events-none"
  class:show
  style="left: {left}px; top: {top}px;"
>
  <button
    type="button"
    class="p-1 rounded bg-white shadow-md hover:bg-slate-50 border border-slate-200 pointer-events-auto"
    title="Spread evenly"
    on:mousedown|stopPropagation
    on:click|stopPropagation={() => dispatch("spreadevenly")}
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
    >
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M11.8637 1.46968C11.5708 1.17679 11.096 1.17679 10.8031 1.46968C10.5102 1.76257 10.5102 2.23745 10.8031 2.53034L12.1894 3.91668L6.00006 3.91667C5.58584 3.91667 5.25006 4.25246 5.25005 4.66667C5.25005 5.08088 5.58584 5.41667 6.00005 5.41667L12.1894 5.41668L10.8031 6.80301C10.5102 7.09591 10.5102 7.57078 10.8031 7.86368C11.0959 8.15657 11.5708 8.15657 11.8637 7.86368L14.5302 5.19721L14.5304 5.19701C14.671 5.05636 14.7501 4.8656 14.7501 4.66668C14.7501 4.46777 14.671 4.27701 14.5304 4.13635L14.5297 4.13566C14.394 4.00034 14.2068 3.91668 14.0001 3.91668L11.8637 1.46968ZM5.19705 8.13634C5.48994 8.42924 5.48994 8.90411 5.19705 9.197L3.81071 10.5833L10 10.5833C10.4143 10.5833 10.75 10.9191 10.75 11.3333C10.75 11.7476 10.4143 12.0833 10 12.0833L3.81071 12.0833L5.19704 13.4697C5.48993 13.7626 5.48993 14.2374 5.19704 14.5303C4.90415 14.8232 4.42927 14.8232 4.13638 14.5303L1.46972 11.8637C1.39781 11.7918 1.34356 11.7089 1.30696 11.6204C1.27029 11.532 1.25005 11.435 1.25005 11.3333C1.25005 11.2316 1.27029 11.1347 1.30696 11.0462C1.34272 10.9598 1.39535 10.8787 1.46484 10.8079M1.47031 10.8024L4.13639 8.13634C4.42928 7.84345 4.90415 7.84345 5.19705 8.13634"
        fill="#64748B"
      />
    </svg>
  </button>
</div>

<style>
  .floating-buttons {
    z-index: 51;
  }

  .floating-buttons.show {
    opacity: 1;
  }
</style>
