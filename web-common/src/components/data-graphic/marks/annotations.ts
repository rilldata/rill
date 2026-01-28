import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
import type {
  GraphicScale,
  SimpleDataGraphicConfiguration,
} from "@rilldata/web-common/components/data-graphic/state/types";
import { Throttler } from "@rilldata/web-common/lib/throttler.ts";
import type { V1MetricsViewAnnotationsResponseAnnotation } from "@rilldata/web-common/runtime-client";
import type { ActionReturn } from "svelte/action";
import { get, writable } from "svelte/store";

export type Annotation = V1MetricsViewAnnotationsResponseAnnotation & {
  startTime: Date;
  endTime?: Date;
  formattedTimeOrRange: string;
};

export type AnnotationGroup = {
  items: Annotation[];
  top: number;
  left: number;
  bottom: number;
  right: number;
  hasRange: boolean;
};

export const AnnotationWidth = 10;
const AnnotationOverlapWidth = AnnotationWidth * (1 - 0.4); // Width where 40% overlap
export const AnnotationHeight = 10;

export class AnnotationsStore {
  public lookupTable = writable<(AnnotationGroup | undefined)[]>([]);
  public annotationGroups = writable<AnnotationGroup[]>([]);
  public hoveredAnnotationGroup = writable<AnnotationGroup | undefined>(
    undefined,
  );

  public annotationPopoverOpened = writable(false);
  public annotationPopoverHovered = writable(false);
  public annotationPopoverTextHiddenCount = writable(0);

  private hoverCheckThrottler = new Throttler(100, 100);

  public updateData(
    annotations: Annotation[],
    scaler: GraphicScale,
    config: SimpleDataGraphicConfiguration,
  ) {
    const groups = this.createAnnotationGroups(annotations, scaler, config);
    this.annotationGroups.set(groups);
    const lookupTable = this.buildLookupTable(groups);
    this.lookupTable.set(lookupTable);
  }

  public triggerHoverCheck(
    mouseoverValue: DomainCoordinates | undefined,
    mouseOverThisChart: boolean,
    annotationPopoverHovered: boolean,
  ) {
    // Check the hover at a slight delay.
    // When the popover is hovered, there will be a small window where both mouseOverThisChart and annotationPopoverHovered will be false.
    this.hoverCheckThrottler.throttle(() =>
      this.checkHover(
        mouseoverValue,
        mouseOverThisChart,
        annotationPopoverHovered,
      ),
    );
  }

  public textHiddenActions = (node: HTMLElement): ActionReturn<void> => {
    let hidden = false;
    const checkTextHidden = () => {
      const currentlyHidden =
        node.scrollWidth > node.clientWidth ||
        node.scrollHeight > node.clientHeight;
      if (currentlyHidden === hidden) return;

      this.annotationPopoverTextHiddenCount.set(
        get(this.annotationPopoverTextHiddenCount) + (currentlyHidden ? 1 : -1),
      );
      hidden = currentlyHidden;
    };
    checkTextHidden();

    node.addEventListener("resize", checkTextHidden);
    return {
      destroy: () => {
        this.annotationPopoverTextHiddenCount.set(
          get(this.annotationPopoverTextHiddenCount) + (hidden ? -1 : 0),
        );
        node.removeEventListener("resize", checkTextHidden);
      },
    };
  };

  private checkHover(
    mouseoverValue: DomainCoordinates | undefined,
    mouseOverThisChart: boolean,
    annotationPopoverHovered: boolean,
  ) {
    const annotationGroups = get(this.annotationGroups);
    const lookupTable = get(this.lookupTable);
    const top = annotationGroups[0]?.top;
    let hoveredAnnotationGroup = get(this.hoveredAnnotationGroup);

    const mouseX = mouseoverValue?.xActual;
    const mouseY = mouseoverValue?.yActual;

    const yNearAnnotations = mouseY !== undefined && mouseY > top;
    const checkXCoord = yNearAnnotations && mouseX !== undefined;

    if (!mouseOverThisChart && !annotationPopoverHovered) {
      // If the mouse is no longer hovering, the current chart or an annotation popover unset the group.
      hoveredAnnotationGroup = undefined;
    } else {
      const tempHoverGroup = checkXCoord ? lookupTable[mouseX] : undefined;
      const hoverGroupChanged =
        tempHoverGroup && tempHoverGroup !== hoveredAnnotationGroup;
      const cursorToLeftOfCurrentGroup =
        hoveredAnnotationGroup &&
        mouseX !== undefined &&
        mouseX < hoveredAnnotationGroup.left;

      if (hoverGroupChanged) {
        // To keep the popover opened for interaction, only update the hovered group when it changes but not when it goes undefined.
        hoveredAnnotationGroup = tempHoverGroup;
      } else if (cursorToLeftOfCurrentGroup) {
        // Else to have better UX, if cursor is to the left of the currently hovered group then unset it.
        hoveredAnnotationGroup = undefined;
      }
    }

    this.hoveredAnnotationGroup.set(hoveredAnnotationGroup);
  }

  private createAnnotationGroups(
    annotations: Annotation[],
    scaler: GraphicScale,
    config: SimpleDataGraphicConfiguration,
  ): AnnotationGroup[] {
    if (annotations.length === 0 || !scaler || !config) return [];

    let currentGroup: AnnotationGroup = this.getSingletonAnnotationGroup(
      annotations[0],
      scaler,
      config,
    );
    const groups: AnnotationGroup[] = [currentGroup];

    for (let i = 1; i < annotations.length; i++) {
      const annotation = annotations[i];
      const group = this.getSingletonAnnotationGroup(
        annotation,
        scaler,
        config,
      );

      const leftDiff = group.left - currentGroup.left;

      if (leftDiff < AnnotationOverlapWidth) {
        currentGroup.right = Math.max(currentGroup.right, group.right);
        currentGroup.items.push(annotation);
      } else {
        currentGroup = group;
        groups.push(currentGroup);
      }
    }

    // Filter out-of-bounds items.
    return groups.filter(
      (g) => g.left > config.plotLeft && g.left < config.plotRight,
    );
  }

  private buildLookupTable(annotationGroups: AnnotationGroup[]) {
    if (annotationGroups.length === 0) return [];
    const lastGroup = annotationGroups[annotationGroups.length - 1];

    const lookupTable = new Array<AnnotationGroup | undefined>(
      Math.ceil(lastGroup.right) + 1,
    ).fill(undefined);

    annotationGroups.forEach((group) => {
      const left = Math.floor(group.left);
      for (let x = 0; x <= AnnotationWidth; x++) {
        lookupTable[left + x] = group;
      }
    });

    return lookupTable;
  }

  private getSingletonAnnotationGroup(
    annotation: Annotation,
    scaler: GraphicScale,
    config: SimpleDataGraphicConfiguration,
  ): AnnotationGroup {
    const left = config.bodyLeft / 2 + scaler(annotation.startTime);
    const right =
      config.bodyLeft / 2 +
      (annotation.endTime
        ? scaler(annotation.endTime)
        : left + AnnotationWidth);
    return <AnnotationGroup>{
      items: [annotation],
      top: config.plotBottom - AnnotationHeight,
      left,
      bottom: config.plotBottom,
      right,
      hasRange: !!annotation.endTime,
    };
  }
}
