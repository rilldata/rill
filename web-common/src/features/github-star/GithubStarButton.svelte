<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import {
    Popover,
    PopoverContent,
  } from "@rilldata/web-common/components/popover";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    GITHUB_STAR_URL,
    githubStarNudge,
    type GithubStarNudge,
  } from "./github-star.svelte";

  /** Injectable so tests can supply a nudge backed by their own store. */
  let { nudge = githubStarNudge }: { nudge?: GithubStarNudge } = $props();

  /** Grace period so the nudge does not appear mid route transition. */
  const AUTO_OPEN_DELAY_MS = 1500;

  // This is a Rill Developer feature only: asking a Cloud viewer or a paying
  // customer to star the repo is off-brand. Today the footer is not mounted on
  // Cloud at all, but that is incidental, so gate it explicitly here rather than
  // relying on a `showFooterLinks={false}` in an unrelated layout.
  const { adminServer } = featureFlags;

  let open = $state(false);
  /**
   * Whether a nudge is outstanding, i.e. the popover opened by itself and the user
   * has not yet answered it. Cleared by the terminal actions so that the close they
   * trigger is not also recorded as a dismissal.
   */
  let nudging = $state(false);
  /** Guards against re-opening unprompted more than once per page load. */
  let autoOpened = false;
  /** The popover has no trigger, so it anchors to the footer link instead. */
  let anchor = $state<HTMLAnchorElement | null>(null);

  // Opening a popover on a timer has no reactive equivalent, so this is one of the
  // cases where an effect is the right tool.
  $effect(() => {
    if ($adminServer || autoOpened || !nudge.visible) return;

    const timeout = setTimeout(() => {
      autoOpened = true;
      nudging = true;
      open = true;
    }, AUTO_OPEN_DELAY_MS);

    return () => clearTimeout(timeout);
  });

  function handleOpenChange(next: boolean) {
    if (next || !nudging) return;
    // Every exit path counts as a soft dismissal: the X, a click outside, or
    // navigating away. They are indistinguishable from here, and not counting
    // silent ignores would re-ask engaged users at every mute expiry.
    nudging = false;
    nudge.recordSoftDismiss();
  }

  // Shared by the footer link and the popover's primary action: both send the user
  // to the repo, so both retire the nudge.
  function star() {
    nudge.recordStar();
    nudging = false;
    open = false;
  }

  function optOut() {
    nudge.recordOptOut();
    nudging = false;
    open = false;
  }
</script>

{#if !$adminServer}
  <!-- A plain link, like the sibling "Report an issue" row: clicking it goes
  straight to the repo rather than opening the popover. -->
  <a
    bind:this={anchor}
    href={GITHUB_STAR_URL}
    target="_blank"
    rel="noreferrer noopener"
    onclick={star}
  >
    <div
      class="flex flex-row items-center px-4 py-1 gap-x-2 text-fg-secondary font-normal hover:bg-popover-accent"
    >
      <!-- Matches the icon sizing workaround in Footer.svelte -->
      <div
        class="grid place-content-center"
        style:width="16px"
        style:height="16px"
      >
        <Github className="fill-fg-secondary" size="14px" />
      </div>
      {m.github_star_footer_label()}
    </div>
  </a>

  <Popover bind:open onOpenChange={handleOpenChange}>
    <!-- Nobody asked for this popover, so it must not take the keyboard. It behaves
    like a notification rather than a dialog: no focus on open, no trap while open,
    and no focus restore on close, since the user may have moved on since. All three
    are needed; bits-ui focuses the content on mount regardless of `trapFocus`, and
    leaving the trap on would drag focus back the moment they clicked away.
    Escape and click-outside listen on the document, so dismissal still works. -->
    <PopoverContent
      customAnchor={anchor}
      align="start"
      side="top"
      sideOffset={8}
      class="github-star-popover w-[280px] overflow-hidden border-primary-200 bg-surface-overlay p-0 shadow-xl"
      role="status"
      trapFocus={false}
      onOpenAutoFocus={(e: Event) => e.preventDefault()}
      onCloseAutoFocus={(e: Event) => e.preventDefault()}
    >
      <div aria-hidden="true" class="h-1 bg-accent-primary"></div>
      <div class="flex flex-col gap-y-4 p-4">
        <div class="flex items-start gap-x-3">
          <div
            class="grid size-9 shrink-0 place-content-center rounded-full bg-primary-50 text-accent-primary-action"
          >
            <Github size="18px" className="fill-current" />
          </div>
          <div class="flex flex-col gap-y-1">
            <h3 class="text-[15px] font-semibold leading-5 text-fg-primary">
              {m.github_star_title()}
            </h3>
            <p class="text-xs leading-4 text-fg-secondary">
              {m.github_star_message()}
            </p>
          </div>
        </div>
        <div class="flex flex-col gap-y-1">
          <Button
            type="primary"
            wide
            href={GITHUB_STAR_URL}
            target="_blank"
            rel="noreferrer noopener"
            onClick={star}
          >
            <Github size="14px" className="fill-current" />
            {m.github_star_cta()}
          </Button>
          <Button type="text" onClick={optOut}>
            {m.github_star_dismiss()}
          </Button>
        </div>
      </div>
    </PopoverContent>
  </Popover>
{/if}

<style>
  @media (prefers-reduced-motion: no-preference) {
    :global(.github-star-popover[data-state="open"]) {
      animation: github-star-popover-in 180ms ease-out;
      transform-origin: bottom left;
    }
  }

  @keyframes github-star-popover-in {
    from {
      opacity: 0;
      transform: translateY(6px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
