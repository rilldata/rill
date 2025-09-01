<!-- @component 
  This component is used to only show a spinner if the loading state is true after a delay.
  This is handy for preventing the spinner from flickering.
-->
<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import DelayedContent from "./DelayedContent.svelte";
  import Spinner from "./Spinner.svelte";

  export let isLoading: boolean;
  export let delay: number = 300;
  export let duration: number = 500;
  export let size: string = "1em";
</script>

<DelayedContent visible={isLoading} {delay}>
  <Spinner status={EntityStatus.Running} {duration} {size} />
</DelayedContent>
