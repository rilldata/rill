import type { Writable } from "svelte/store";
import { PivotChipType } from "./types";

type Options = {
  condition: boolean;
  ghostIndex: Writable<number | null>;
  chipType: PivotChipType | undefined;
  canMixTypes: boolean;
  orientation?: "horizontal" | "vertical";
};

export function swapListener(
  node: HTMLElement,
  {
    condition,
    ghostIndex,
    chipType,
    canMixTypes,
    orientation = "horizontal",
  }: Options,
) {
  let added = false;
  const children = node.getElementsByClassName(
    "drag-item",
  ) as HTMLCollectionOf<HTMLDivElement>;
  const mousemoveHandler = (
    e: MouseEvent & { currentTarget: HTMLDivElement },
  ) => handlePillShift(e, orientation);

  function updateListener({ condition, chipType, canMixTypes }: Options) {
    if (condition && !added) {
      if (canMixTypes !== true) {
        const category =
          chipType === PivotChipType.Measure ? "measure" : "dimension";
        for (const child of children) {
          if (category !== child.dataset.type) continue;
          child.addEventListener("mousemove", mousemoveHandler);
        }
      } else {
        for (const child of children) {
          child.addEventListener("mousemove", mousemoveHandler);
        }
      }
      added = true;
    } else if (!condition && added) {
      for (const child of children) {
        child.removeEventListener("mousemove", mousemoveHandler);
      }
      added = false;
    }
  }

  function handlePillShift(
    e: MouseEvent & { currentTarget: HTMLDivElement },
    orientation: Options["orientation"],
  ) {
    const index = Number(e.currentTarget.dataset.index);

    if (orientation === "vertical") {
      const { height, top } = e.currentTarget.getBoundingClientRect();
      const midwayPoint = top + height / 2;
      const isTop = e.clientY <= midwayPoint;
      const newIndex = isTop ? index : index + 1;
      ghostIndex.set(newIndex);
    } else {
      const { width, left } = e.currentTarget.getBoundingClientRect();
      const midwayPoint = left + width / 2;

      const isLeft = e.clientX <= midwayPoint;

      const newIndex = isLeft ? index : index + 1;

      ghostIndex.set(newIndex);
    }
  }

  updateListener({ condition, ghostIndex, chipType, canMixTypes });

  return {
    update({ condition, chipType, ghostIndex, canMixTypes }: Options) {
      updateListener({ condition, ghostIndex, chipType, canMixTypes });
    },

    destroy() {
      for (const child of children) {
        child.removeEventListener("mousemove", mousemoveHandler);
      }
    },
  };
}
