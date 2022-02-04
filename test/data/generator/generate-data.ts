import "module-alias/register";
import {DataGeneratorFarm} from "./DataGeneratorFarm";

(async () => {
    const dataGeneratorFarm = new DataGeneratorFarm(__dirname + "/generate-data-worker");
    await dataGeneratorFarm.generate(process.argv[2], Number(process.argv[3]));
    dataGeneratorFarm.stop();
})();
