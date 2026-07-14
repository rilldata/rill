import { writable } from "svelte/store";
import { findHoveredGroup, type AnnotationGroup } from "./annotation-utils";

const POPOVER_DELAY_MS = 150;

/**
 * Manages annotation popover hover state: hit-testing mouse position against
 * annotation groups, delayed hiding so the user can reach the popover, and
 * popover-is-hovered tracking.
 */
export class AnnotationPopoverController {
  readonly hoveredGroup = writable<AnnotationGroup | null>(null);

  private popoverHovered = false;
  private timeout: ReturnType<typeof setTimeout> | null = null;
  private currentGroup: AnnotationGroup | null = null;

  /** Call from SVG mousemove. */
  checkHover(e: MouseEvent, groups: AnnotationGroup[], isScrubbing: boolean) {
    if (isScrubbing || groups.length === 0) {
      this.scheduleClear();
      return;
    }
    const svg = e.currentTarget as SVGSVGElement;
    const rect = svg.getBoundingClientRect();
    const hit = findHoveredGroup(
      groups,
      e.clientX - rect.left,
      e.clientY - rect.top,
    );
    if (hit) {
      this.cancelTimeout();
      this.setGroup(hit);
    } else if (this.currentGroup && !this.popoverHovered) {
      this.scheduleClear();
    }
  }

  /** Call when the popover itself is hovered / unhovered. */
  setPopoverHovered(hovered: boolean) {
    this.popoverHovered = hovered;
    this.cancelTimeout();
    if (!hovered) this.scheduleClear();
  }

  /** Schedule a delayed clear (e.g. on mouseleave). */
  scheduleClear() {
    if (this.popoverHovered || this.timeout) return;
    this.timeout = setTimeout(() => {
      if (!this.popoverHovered) this.setGroup(null);
      this.timeout = null;
    }, POPOVER_DELAY_MS);
  }

  destroy() {
    this.cancelTimeout();
  }

  private setGroup(group: AnnotationGroup | null) {
    this.currentGroup = group;
    this.hoveredGroup.set(group);
  }

  private cancelTimeout() {
    if (this.timeout) {
      clearTimeout(this.timeout);
      this.timeout = null;
    }
  }
}
