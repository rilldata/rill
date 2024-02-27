<script lang="ts">
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { fly } from "svelte/transition";
  import Spinner from "../features/entity-management/Spinner.svelte";
  import { EntityStatus } from "../features/entity-management/types";

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
    500,
  );
</script>

<Overlay {bg}>
  <div
    transition:fly|global={{ duration: 200, y: 16 }}
    class="text-white text-center flex flex-col gap-y-4"
    style:width="540px"
  >
    <div class="flex flex-col gap-y-3">
      <div
        class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all"
        style:font-size="48px"
      >
        <div class="on" style="--length: {2000 + Math.random() * 5000}ms;">
          <Spinner {status} duration={300 + Math.random() * 200} />
        </div>
      </div>
      <slot name="title" />
    </div>
    <slot name="detail" />
  </div>
</Overlay>
