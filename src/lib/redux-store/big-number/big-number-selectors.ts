import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";

export const {
  singleSelector: selectBigNumberById,
  manySelectorByIds: selectBigNumbersByIds,
} = generateEntitySelectors<BigNumberEntity>("bigNumber");
