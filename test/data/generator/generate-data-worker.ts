import "module-alias/register";
import {DataGeneratorWorker} from "./DataGeneratorWorker";

const dataGeneratorWorker = new DataGeneratorWorker();

module.exports.generate = function updateNewCases(
    [type, startId]: [string, number], callback: (err: Error, rows: Array<Record<string, any>>) => void,
) {
    callback(null, dataGeneratorWorker.generate(type, startId));
}
