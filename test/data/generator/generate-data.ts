import "module-alias/register";
import {DataGeneratorFarm} from "./DataGeneratorFarm";
import {AD_BID_COUNT, AD_IMPRESSION_COUNT, MAX_USERS} from "./data-constants";

const generators: Array<[string, number]> = [
    ["AdBids", AD_BID_COUNT],
    ["AdImpressions", AD_IMPRESSION_COUNT],
    ["Users", MAX_USERS],
];

(async () => {
    const dataGeneratorFarm = new DataGeneratorFarm(__dirname + "/generate-data-worker");
    for (const generator of generators) {
        await dataGeneratorFarm.generate(generator[0], generator[1]);
    }
    dataGeneratorFarm.stop();
})();
