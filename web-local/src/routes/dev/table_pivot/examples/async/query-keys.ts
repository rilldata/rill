import type { PivotPos } from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  get2DBlocks,
  getBlocks,
} from "@rilldata/web-common/features/dashboards/time-dimension-details/util";

const ROW_BLOCK_SIZE = 50;
const COL_BLOCK_SIZE = 50;

export function getRowHeaderKeysFromPos(pos: PivotPos, config: any) {
  const blocks = getBlocks(ROW_BLOCK_SIZE, pos.y0, pos.y1);
  return blocks.map((b) => ["async-pivot-row-header", config, b[0], b[1]]);
}

export function getColHeaderKeysFromPos(pos: PivotPos, config: any) {
  const blocks = getBlocks(COL_BLOCK_SIZE, pos.x0, pos.x1);
  // make the key exclude stuff it doesnt care about, like expanded state
  return blocks.map((b) => ["async-pivot-col-header", config, b[0], b[1]]);
}

export function getBodyKeysFromPos(pos: PivotPos, config: any) {
  const blocks = get2DBlocks({
    blockSizeX: COL_BLOCK_SIZE,
    blockSizeY: ROW_BLOCK_SIZE,
    ...pos,
  });
  return blocks.map((b) => ["async-pivot-body", config, b]);
}
