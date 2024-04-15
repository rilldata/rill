import { PivotChipType } from "./types";
import type { Writable } from "svelte/store";

type Options = {
  condition: boolean;
  ghostIndex: Writable<number | null>;
  chipType: PivotChipType | undefined;
};

export function swapListener(
  node: HTMLElement,
  { condition, ghostIndex, chipType }: Options,
) {
  let added = false;
  const children = node.getElementsByClassName(
    "drag-item",
  ) as HTMLCollectionOf<HTMLDivElement>;

  function updateListener({ condition, chipType }: Options) {
    if (condition && !added) {
      const category =
        chipType === PivotChipType.Measure ? "measure" : "dimension";
      for (const child of children) {
        if (category !== child.dataset.type) continue;
        child.addEventListener("mousemove", handlePillShift);
      }
      added = true;
    } else if (!condition && added) {
      for (const child of children) {
        child.removeEventListener("mousemove", handlePillShift);
      }
      added = false;
    }
  }

  function handlePillShift(e: MouseEvent & { currentTarget: HTMLDivElement }) {
    const index = Number(e.currentTarget.dataset.index);

    const { width, left } = e.currentTarget.getBoundingClientRect();
    const midwayPoint = left + width / 2;

    const isLeft = e.clientX <= midwayPoint;

    const newIndex = isLeft ? index : index + 1;

    ghostIndex.set(newIndex);
  }

  updateListener({ condition, ghostIndex, chipType });

  return {
    update({ condition, chipType, ghostIndex }: Options) {
      updateListener({ condition, ghostIndex, chipType });
    },

    destroy() {
      for (const child of children) {
        child.removeEventListener("mousemove", handlePillShift);
      }
    },
  };
}
