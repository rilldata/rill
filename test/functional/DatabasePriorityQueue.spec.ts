import { FunctionalTestBase } from "./FunctionalTestBase";
import type { SinonSpy } from "sinon";
import { SingleTableQuery, TwoTableJoinQuery } from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";
import {
    AdBidsImportActions,
    AdBidsProfilingActions,
    SingleQueryProfilingActions, TwoTableJoinQueryProfilingActions
} from "../data/DatabasePriorityQueue.data";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

@FunctionalTestBase.Suite
export class DatabasePriorityQueueSpec extends FunctionalTestBase {
    private databaseDispatchSpy: SinonSpy;

    public async setup() {
        await super.setup();

        this.databaseDispatchSpy = this.sandbox.spy(
            this.serverDataModelerService.getDatabaseService(), "dispatch");
    }

    @FunctionalTestBase.BeforeEachTest()
    public async setupTests() {
        await this.clientDataModelerService.dispatch("clearAllTables", []);
        await this.clientDataModelerService.dispatch("clearAllModels", []);
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_0", query: ""}]);
    }

    @FunctionalTestBase.Test()
    public async shouldDePrioritiseTableProfiling() {
        const importPromise = this.clientDataModelerService.dispatch(
            "addOrUpdateTableFromFile", ["data/AdBids.parquet"]);
        await asyncWait(1);

        const [model] = this.getModels("tableName", "query_0");
        const modelQueryPromise = this.clientDataModelerService.dispatch(
            "updateModelQuery", [model.id, SingleTableQuery]);

        await Promise.all([importPromise, modelQueryPromise]);
        expect(this.databaseDispatchSpy.args).toStrictEqual([
            ...AdBidsImportActions,
            // this is where the switch to model happens
            [
                "createViewOfQuery",
                [
                    "query_0",
                    "select count(*) as impressions,publisher,domain from 'AdBids' group by publisher,domain"
                ]
            ],
            // there are some AdBids actions mixed in
            // because of some code running within ModelActions.collectModelInfo
            [ "getNumericHistogram", [ "AdBids", "id", "INTEGER" ] ],
            [ "getProfileColumns", [ "query_0" ] ],
            [ "getDescriptiveStatistics", [ "AdBids", "id" ] ],
            ...SingleQueryProfilingActions,
            // table profiling continues here
            ...AdBidsProfilingActions,
        ]);
    }

    @FunctionalTestBase.Test()
    public async shouldStopOlderQueriesOfModel() {
        await this.clientDataModelerService.dispatch(
            "addOrUpdateTableFromFile", ["data/AdBids.parquet"]);
        await this.clientDataModelerService.dispatch(
            "addOrUpdateTableFromFile", ["data/AdImpressions.parquet"]);
        this.databaseDispatchSpy.resetHistory();

        const [model] = this.getModels("tableName", "query_0");
        const modelQueryOnePromise = this.clientDataModelerService.dispatch(
            "updateModelQuery", [model.id, TwoTableJoinQuery]);
        await asyncWait(100);
        const modelQueryTwoPromise = this.clientDataModelerService.dispatch(
            "updateModelQuery", [model.id, SingleTableQuery]);

        await Promise.all([modelQueryOnePromise, modelQueryTwoPromise]);
        expect(this.databaseDispatchSpy.args).toStrictEqual([
            [
                "validateQuery",
                [
                    "\n" +
                    "select count(*) as impressions, avg(bid.bid_price) as bid_price, bid.publisher, bid.domain, imp.city, imp.country\n" +
                    "from 'AdBids' bid join 'AdImpressions' imp on bid.id = imp.id\n" +
                    "group by bid.publisher, bid.domain, imp.city, imp.country\n"
                ]
            ],
            [
                "createViewOfQuery",
                [
                    "query_0",
                    "select count(*) as impressions,avg(bid.bid_price) as bid_price,bid.publisher,bid.domain,imp.city,imp.country from 'AdBids' bid join 'AdImpressions' imp on bid.id = imp.id group by bid.publisher,bid.domain,imp.city,imp.country"
                ]
            ],
            [ "getProfileColumns", [ "query_0" ] ],
            [ "getNumericHistogram", [ "query_0", "impressions", "BIGINT" ] ],
            [
                "validateQuery",
                [
                    "select count(*) as impressions, publisher, domain from 'AdBids' group by publisher, domain"
                ]
            ],
            [
                "createViewOfQuery",
                [
                    "query_0",
                    "select count(*) as impressions,publisher,domain from 'AdBids' group by publisher,domain"
                ]
            ],
            [ "getProfileColumns", [ "query_0" ] ],
            ...SingleQueryProfilingActions,
        ]);
    }

    @FunctionalTestBase.Test()
    public async shouldDePrioritiseInactiveModel() {
        await this.clientDataModelerService.dispatch(
            "addOrUpdateTableFromFile", ["data/AdBids.parquet"]);
        await this.clientDataModelerService.dispatch(
            "addOrUpdateTableFromFile", ["data/AdImpressions.parquet"]);
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_1", query: ""}]);
        this.databaseDispatchSpy.resetHistory();

        const [model0] = this.getModels("tableName", "query_1");
        const modelQueryOnePromise = this.clientDataModelerService.dispatch(
            "updateModelQuery", [model0.id, TwoTableJoinQuery]);
        await this.clientDataModelerService.dispatch("setActiveAsset",
            [EntityType.Model, model0.id]);
        await asyncWait(50);
        const [model1] = this.getModels("tableName", "query_0");
        const modelQueryTwoPromise = this.clientDataModelerService.dispatch(
            "updateModelQuery", [model1.id, SingleTableQuery]);
        await asyncWait(50);
        await this.clientDataModelerService.dispatch("setActiveAsset",
            [EntityType.Model, model1.id]);

        await Promise.all([modelQueryOnePromise, modelQueryTwoPromise]);
        expect(this.databaseDispatchSpy.args).toStrictEqual([
            [
                "validateQuery",
                [
                    "\n" +
                    "select count(*) as impressions, avg(bid.bid_price) as bid_price, bid.publisher, bid.domain, imp.city, imp.country\n" +
                    "from 'AdBids' bid join 'AdImpressions' imp on bid.id = imp.id\n" +
                    "group by bid.publisher, bid.domain, imp.city, imp.country\n"
                ]
            ],
            [
                "createViewOfQuery",
                [
                    "query_1",
                    "select count(*) as impressions,avg(bid.bid_price) as bid_price,bid.publisher,bid.domain,imp.city,imp.country from 'AdBids' bid join 'AdImpressions' imp on bid.id = imp.id group by bid.publisher,bid.domain,imp.city,imp.country"
                ]
            ],
            [ "getProfileColumns", [ "query_1" ] ],
            [ "getNumericHistogram", [ "query_1", "impressions", "BIGINT" ] ],
            [
                "validateQuery",
                [
                    "select count(*) as impressions, publisher, domain from 'AdBids' group by publisher, domain"
                ]
            ],
            [ "getDescriptiveStatistics", [ "query_1", "impressions" ] ],
            [
                "createViewOfQuery",
                [
                    "query_0",
                    "select count(*) as impressions,publisher,domain from 'AdBids' group by publisher,domain"
                ]
            ],
            [ "getNullCount", [ "query_1", "impressions" ] ],
            [ "getProfileColumns", [ "query_0" ] ],
            [ "getNumericHistogram", [ "query_1", "bid_price", "DOUBLE" ] ],
            // single table query (query_0) is executed 1st.
            ...SingleQueryProfilingActions,
            // two table join query (query_0) is executed next.
            ...TwoTableJoinQueryProfilingActions,
        ]);
    }
}
