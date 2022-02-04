import {AdBidsGeneratorType} from "./AdBidsGeneratorType";
import type {DataGeneratorType} from "./DataGeneratorType";

export const BATCH_SIZE = 100;
export const DATA_GENERATOR_TYPE_MAP: {
    [type in string]: DataGeneratorType
} = {
    "AdBids": new AdBidsGeneratorType(),
};
