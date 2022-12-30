<script lang="ts">
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { fly } from "svelte/transition";
  import Spinner from "../../../lib/components/Spinner.svelte";
  import { EntityStatus } from "../../../lib/temp/entity";

  export let bg = "rgba(0,0,0,.6)";

  let status = EntityStatus.Running;

  setTimeout(
    () =>
      setInterval(() => {
        status =
          status === EntityStatus.Running
            ? EntityStatus.Idle
            : EntityStatus.Running;
      }, 1000),
    500
  );
</script>

<Overlay {bg}>
  <div
    transition:fly={{ duration: 200, y: 16 }}
    class="text-white text-center"
    style:width="540px"
  >
    <div
      class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all"
      style:font-size="48px"
    >
      <div class="on" style="--length: {2000 + Math.random() * 5000}ms;">
        <Spinner {status} bg="white" duration={300 + Math.random() * 200} />
      </div>
    </div>
    <div class=" pt-3"><slot name="title" /></div>
    <div>
      <slot name="text" />
    </div>
  </div>
</Overlay>
