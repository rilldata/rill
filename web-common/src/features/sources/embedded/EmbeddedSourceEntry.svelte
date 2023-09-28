<!-- @component a display element for a single embedded source, used in the navigation and inspector.

-->
<script lang="ts">
  import { truncateMiddleText } from "@rilldata/web-common/lib/actions/truncate-middle-text";
  import ConnectorLabel from "./ConnectorLabel.svelte";

  // TODO: remove embedded support

  export let connector: string;
  export let path: string;

  function cutOutProtocol(url: string) {
    if (url.startsWith("https://")) return url.slice("https://".length);
    if (url.startsWith("gs://")) return url.slice("gs://".length);
    if (url.startsWith("s3://")) return url.slice("s3://".length);
    return url;
  }
</script>

<div class="w-full overflow-x-hidden flex items-center gap-x-2">
  <div>
    <ConnectorLabel {connector} />
  </div>
  <div class="w-full overflow-x-hidden">
    <div
      aria-label={path}
      class=" overflow-hidden whitespace-nowrap"
      style:min-width="52px"
      use:truncateMiddleText
    >
      {cutOutProtocol(path)}
    </div>
  </div>
</div>
