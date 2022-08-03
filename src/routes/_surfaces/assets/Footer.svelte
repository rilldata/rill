<script lang="ts">
  import Discord from "$lib/components/icons/Discord.svelte";
  import Docs from "$lib/components/icons/Docs.svelte";
  import InfoCircle from "$lib/components/icons/InfoCircle.svelte";
  import { getContext } from "svelte";

  const metadata = getContext("rill:app:metadata");

  const lineItems = [
    {
      icon: Docs,
      label: "Docs",
      href: "https://docs.rilldata.com",
      className: "fill-gray-600",
    },
    {
      icon: Discord,
      label: "Share your feedback",
      href: "http://bit.ly/3jg4IsF",
      className: "fill-gray-500",
    },
  ];
</script>

<div class="flex flex-col bg-gray-50 pt-3 pb-1 gap-y-1">
  {#each lineItems as lineItem}
    <a href={lineItem.href} target="_blank"
      ><div
        class="flex flex-row items-center px-4 py-1 gap-x-2 text-gray-700 font-bold hover:bg-gray-200"
      >
        <div>
          <svelte:component
            this={lineItem.icon}
            className={lineItem.className}
            size="1.25em"
          />
        </div>
        {lineItem.label}
      </div></a
    >
  {/each}
  <div
    class="italic px-4 py-1 text-gray-600 flex flex-row items-center gap-x-2"
    style:font-size="10px"
  >
    <span class="text-gray-400">
      <InfoCircle size="16px" />
    </span>
    version {metadata?.version ?? "unknown version"} –
    {metadata?.commit ?? "no commit"}
  </div>
</div>
