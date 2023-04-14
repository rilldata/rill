<script lang="ts">
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";

  export let organization: string;
  export let project: string;
</script>

<div
  class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto"
  style="max-width: 400px"
>
  <CancelCircleInverse size="7em" className="text-gray-200" />
  <h1 class="text-xl font-normal text-gray-800">
    Your dashboard isn't working right now.
  </h1>
  <button
    class="border border-blue-300 text-blue-600 w-full h-10 hover:bg-slate-100 hover:border-gray-100"
    on:click={() => {
      window.location.href = `/${organization}/${project}/-/deployment`;
    }}
    >View project status
  </button>
  <p class="text-gray-500">
    Need help? Reach out to us on <a href="http://bit.ly/3jg4IsF">Discord</a>
  </p>
</div>
