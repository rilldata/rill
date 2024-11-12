<script lang="ts">
  import { timeAgo } from "../../dashboards/listing/utils";

  export let createdOn: string;
  export let updatedOn: string;

  function formatText() {
    if (updatedOn) {
      return `Updated ${timeAgo(new Date(updatedOn))}`;
    }
    return `Added ${timeAgo(new Date(createdOn))}`;
  }
  $: activityText = formatText();
</script>

<span
  class="text-xs text-gray-500"
  title={new Date(updatedOn || createdOn).toLocaleString()}
>
  {activityText}
</span>
