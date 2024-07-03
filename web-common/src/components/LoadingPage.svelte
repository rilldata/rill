<script lang="ts">
  import Spinner from "../features/entity-management/Spinner.svelte";
  import { EntityStatus } from "../features/entity-management/types";
  import CtaContentContainer from "./calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "./calls-to-action/CTALayoutContainer.svelte";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36 mt-10">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
