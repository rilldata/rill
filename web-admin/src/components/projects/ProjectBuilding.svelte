<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let organization: string;
  export let project: string;
</script>

<div
  class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto"
  style="max-width: 400px"
>
  <div class="h-36">
    <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
  </div>
  <h1 class="text-xl font-normal text-gray-800">
    Hang tight! We're building your dashboard...
  </h1>
  <button
    class="border border-blue-300 text-blue-600 w-full h-10 hover:bg-slate-100 hover:border-gray-100"
    on:click={() => {
      window.location.href = `/${organization}/${project}`;
    }}
    >View project status
  </button>
  <p class="text-gray-500">
    Need help? Reach out to us on <a href="http://bit.ly/3jg4IsF">Discord</a>
  </p>
</div>
