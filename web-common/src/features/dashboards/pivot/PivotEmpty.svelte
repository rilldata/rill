<script>
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import EmptyTableIcon from "./EmptyTableIcon.svelte";

  export let isFetching = false;
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-6">
  {#if isFetching}
    <Spinner size="64px" status={EntityStatus.Running} />
    <div class="font-semibold text-gray-800 mt-1 text-lg">
      Hang tight! We're building your pivot table...
    </div>
    <div class="text-gray-600">
      Need help? Reach out to us on <a
        target="_blank"
        href="http://bit.ly/3jg4IsF">Discord</a
      >
    </div>
  {:else}
    <EmptyTableIcon />
    <div class="flex flex-col items-center gap-y-2">
      <div class="font-semibold text-gray-800 mt-1 text-lg">
        Your pivot table looks lonely
      </div>
      <div class="text-gray-600 text-base">
        Give it some data to keep it company.
      </div>
    </div>
    <div class="text-gray-600">
      Learn more about pivot tables in our <a
        target="_blank"
        href="https://docs.rilldata.com/">docs</a
      >.
    </div>
  {/if}
</div>
