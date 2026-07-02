<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<div
  class="m-auto flex-col justify-center items-center gap-y-4 inline-flex w-full"
>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-fg-secondary font-semibold">
      {m.public_url_no_urls_title()}
    </div>
    <div>
      <span style="text-fg-secondary font-normal">
        {m.public_url_no_urls_description()}
      </span>
    </div>
  </div>
</div>
