import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";

/**
 * select
 * count(*), date_trunc('HOUR', created_date) as inter
 * from nyc311_reduced
 * group by date_trunc('HOUR', created_date) order by inter;
 */

export class DimensionsActions extends RillDeveloperActions {
  public async inferDimension() {}
}
