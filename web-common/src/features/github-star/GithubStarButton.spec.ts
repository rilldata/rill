import { render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { InMemoryRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import GithubStarButton from "./GithubStarButton.svelte";
import {
  GITHUB_STAR_URL,
  GithubStarNudge,
  type GithubStarState,
} from "./github-star.svelte";

/** Lets the auto-open timer and the resulting Svelte update settle. */
function settle() {
  return vi.advanceTimersByTimeAsync(2000);
}

/** A nudge of its own, so these tests neither touch localStorage nor the app-wide singleton. */
function createNudge() {
  return new GithubStarNudge(
    new InMemoryRuneStore<GithubStarState>({ status: "unarmed" }),
  );
}

function renderArmed() {
  const nudge = createNudge();
  nudge.armPayoff();
  render(GithubStarButton, { props: { nudge } });
  return nudge;
}

const outsideElements: HTMLElement[] = [];

/** Stands in for whatever the user was actually working on, e.g. the editor. */
function renderOutsideElement() {
  const element = document.createElement("button");
  document.body.append(element);
  outsideElements.push(element);
  return element;
}

describe("GithubStarButton", () => {
  beforeEach(() => {
    featureFlags.adminServer.resetToDefault();
    vi.useFakeTimers();
  });

  afterEach(() => {
    featureFlags.adminServer.resetToDefault();
    outsideElements.splice(0).forEach((element) => element.remove());
    vi.useRealTimers();
  });

  it("renders nothing at all on Rill Cloud", async () => {
    featureFlags.adminServer.set(true);
    const nudge = renderArmed();
    await settle();

    // Neither the button nor the nudge: this is a Rill Developer feature.
    expect(screen.queryByText("Star us on GitHub")).not.toBeInTheDocument();
    expect(screen.queryByText("Enjoying Rill?")).not.toBeInTheDocument();
    expect(nudge.state.status).toBe("armed");
    expect(nudge.state.mutedUntil).toBeUndefined();
  });

  it("always renders the footer link, pointing straight at the repo", () => {
    render(GithubStarButton, { props: { nudge: createNudge() } });

    const link = screen.getByText("Star us on GitHub").closest("a");
    expect(link).toHaveAttribute("href", GITHUB_STAR_URL);
    expect(link).toHaveAttribute("target", "_blank");
  });

  it("does not open unprompted when no payoff has happened", async () => {
    render(GithubStarButton, { props: { nudge: createNudge() } });
    await settle();

    expect(screen.queryByText("Enjoying Rill?")).not.toBeInTheDocument();
  });

  it("opens unprompted once a payoff has armed it", async () => {
    renderArmed();
    await settle();

    expect(screen.getByText("Enjoying Rill?")).toBeInTheDocument();
  });

  it("does not take the keyboard when the nudge opens", async () => {
    const editor = renderOutsideElement();
    editor.focus();

    renderArmed();
    await settle();

    expect(screen.getByText("Enjoying Rill?")).toBeInTheDocument();
    expect(document.activeElement).toBe(editor);
  });

  it("lets focus leave the open nudge instead of trapping it", async () => {
    const editor = renderOutsideElement();
    renderArmed();
    await settle();

    editor.focus();
    await settle();

    // The nudge stays open; it just does not hold the keyboard hostage.
    expect(screen.getByText("Enjoying Rill?")).toBeInTheDocument();
    expect(document.activeElement).toBe(editor);
  });

  it("leaves focus where it is when the nudge closes", async () => {
    renderArmed();
    await settle();

    const editor = renderOutsideElement();
    editor.focus();
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    await settle();

    expect(screen.queryByText("Enjoying Rill?")).not.toBeInTheDocument();
    expect(document.activeElement).toBe(editor);
  });

  it("counts dismissing the nudge with Escape as a soft dismissal", async () => {
    const nudge = renderArmed();
    await settle();

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    await settle();

    // Still armed, but muted: the ask is deferred rather than abandoned.
    expect(nudge.state.status).toBe("armed");
    expect(nudge.state.mutedUntil).toBeGreaterThan(Date.now());
  });

  it("retires the nudge when the footer link is clicked, without opening the popover", async () => {
    // Armed, and clicked inside the grace period: the pending auto-open must be
    // cancelled rather than firing at someone already on their way to the repo.
    const nudge = renderArmed();
    screen.getByText("Star us on GitHub").closest("a")!.click();
    await settle();

    expect(screen.queryByText("Enjoying Rill?")).not.toBeInTheDocument();
    expect(nudge.state.status).toBe("done");
    expect(nudge.state.mutedUntil).toBeUndefined();
  });

  it("records an opt-out as terminal without spending a dismissal", async () => {
    const nudge = renderArmed();
    await settle();

    screen.getByText("Don't show again").click();
    await settle();

    expect(nudge.state.status).toBe("done");
    expect(nudge.state.mutedUntil).toBeUndefined();
  });

  it("records the star click as terminal without spending a dismissal", async () => {
    const nudge = renderArmed();
    await settle();

    screen.getByText("Star on GitHub").click();
    await settle();

    expect(nudge.state.status).toBe("done");
    expect(nudge.state.mutedUntil).toBeUndefined();
  });

  it("does not re-open unprompted after being dismissed", async () => {
    renderArmed();
    await settle();

    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    await settle();
    await settle();

    expect(screen.queryByText("Enjoying Rill?")).not.toBeInTheDocument();
  });
});
