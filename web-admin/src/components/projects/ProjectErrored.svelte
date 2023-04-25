<script lang="ts">
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto">
  <CancelCircleInverse size="7em" className="text-gray-200" />
  <div class="flex flex-col items-center space-y-2">
    <h1 class="text-lg font-semibold">
      Sorry, your dashboard isn't working right now!
    </h1>
    <p class="text-gray-500 text-base">
      View project status for errors or contact your organization's admin.
    </p>
  </div>
  <button
    class="border border-blue-300 text-blue-600 w-full h-10 hover:bg-slate-100 hover:border-gray-100 max-w-[400px]"
    on:click={() => {
      window.location.href = `/${organization}/${project}/-/deployment`;
    }}
    >View project status
  </button>
  <p class="text-gray-500">
    Need help? Reach out to us on <a href="http://bit.ly/3jg4IsF">Discord</a>
  </p>
</div>
