import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
import type {
  GraphicScale,
  SimpleDataGraphicConfiguration,
} from "@rilldata/web-common/components/data-graphic/state/types";
import { Throttler } from "@rilldata/web-common/lib/throttler.ts";
import { get, writable } from "svelte/store";

export type Annotation = {
  startTime: Date;
  endTime?: Date;
  formattedTimeOrRange: string;
  grain?: string;
  description: string;
};

export type AnnotationGroup = {
  items: Annotation[];
  top: number;
  left: number;
  bottom: number;
  right: number;
  hasRange: boolean;
};

// h-[24px] w-[12px]
export const AnnotationWidth = 12;
const AnnotationOverlapWidth = AnnotationWidth * (1 - 0.66); // Width where 66% overlap
export const AnnotationHeight = 10;

export class AnnotationsStore {
  public lookupTable = writable<(AnnotationGroup | undefined)[]>([]);
  public annotationGroups = writable<AnnotationGroup[]>([]);
  public hoveredAnnotationGroup = writable<AnnotationGroup | undefined>(
    undefined,
  );

  public annotationPopoverOpened = writable<boolean>(false);
  public annotationPopoverHovered = writable<boolean>(false);

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
    this.hoverCheckThrottler.throttle(() =>
      this.checkHover(
        mouseoverValue,
        mouseOverThisChart,
        annotationPopoverHovered,
      ),
    );
  }

  private checkHover(
    mouseoverValue: DomainCoordinates | undefined,
    mouseOverThisChart: boolean,
    annotationPopoverHovered: boolean,
  ) {
    const annotationGroups = get(this.annotationGroups);
    const lookupTable = get(this.lookupTable);
    const hovered = mouseOverThisChart || annotationPopoverHovered;
    const top = annotationGroups[0]?.top;

    const mouseX = mouseoverValue?.xActual;
    const mouseY = mouseoverValue?.yActual;

    const yNearAnnotations = mouseY !== undefined && mouseY > top;
    const checkXCoord = yNearAnnotations && mouseX !== undefined;

    let hoveredAnnotationGroup = get(this.hoveredAnnotationGroup);

    if (!hovered) {
      hoveredAnnotationGroup = undefined;
    } else {
      const tempHoveredAnnotationGroup = checkXCoord
        ? lookupTable[mouseX]
        : undefined;
      if (
        tempHoveredAnnotationGroup &&
        tempHoveredAnnotationGroup !== hoveredAnnotationGroup
      ) {
        hoveredAnnotationGroup = tempHoveredAnnotationGroup;
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

    return groups;
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
    const left = scaler(annotation.startTime);
    const right = annotation.endTime
      ? scaler(annotation.endTime)
      : left + AnnotationWidth;
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
