import { AdBidsGeneratorType } from "./AdBidsGeneratorType";
import type { DataGeneratorType } from "./DataGeneratorType";
import { UsersGeneratorType } from "./UsersGeneratorType";
import { AdImpressionsGeneratorType } from "./AdImpressionsGeneratorType";

export const DATA_GENERATOR_TYPE_MAP: {
  [type in string]: DataGeneratorType;
} = {
  AdBids: new AdBidsGeneratorType(),
  AdImpressions: new AdImpressionsGeneratorType(),
  Users: new UsersGeneratorType(),
};
