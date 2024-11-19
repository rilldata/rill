<script lang="ts">
  import { timeAgo } from "../../dashboards/listing/utils";

  export let createdOn: string;
  export let updatedOn: string;

  function formatText() {
    const updatedDate = new Date(updatedOn);

    // FIXME: `updateProjectVariables` does not update the `updatedOn` timestamp correctly
    // if (createdDate.getTime() === updatedDate.getTime()) {
    //   return `Added ${timeAgo(createdDate)}`;
    // }

    return `Updated ${timeAgo(updatedDate)}`;
  }

  $: formattedText = formatText();
</script>

<span
  class="text-xs text-gray-500"
  title={new Date(updatedOn || createdOn).toLocaleString()}
>
  {formattedText}
</span>
