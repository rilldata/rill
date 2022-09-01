package com.rilldata.calcite.visitors;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.extensions.SqlCreateMetric;
import com.rilldata.calcite.models.Artifact;
import com.rilldata.calcite.models.ArtifactManager;
import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.SqlOrderBy;
import org.apache.calcite.sql.SqlSelect;
import org.apache.calcite.sql.SqlWith;
import org.apache.calcite.sql.SqlWithItem;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.sql.util.SqlBasicVisitor;
import org.apache.calcite.sql.validate.SqlValidatorException;

import java.util.Map;

/**
 * This is the main visitor used to expand user query as per the METRICS VIEW model
 * */
public class MetricsViewExpander extends SqlBasicVisitor<SqlNode>
{
  ArtifactManager artifactManager;
  CalciteToolbox calciteToolbox;

  public MetricsViewExpander(ArtifactManager artifactManager, CalciteToolbox calciteToolbox)
  {
    this.artifactManager = artifactManager;
    this.calciteToolbox = calciteToolbox;
  }

  @Override public SqlNode visit(SqlCall sqlCall)
  {
    if (sqlCall.getKind().equals(SqlKind.ORDER_BY)) {
      // its a limit query, extract the select query and recursively expand that
      ((SqlOrderBy) sqlCall).query.accept(this);
      return sqlCall;
    } else if (sqlCall.getKind().equals(SqlKind.WITH)) {
      // CTEs
      SqlWith sqlWith = (SqlWith) sqlCall;
      for (SqlNode sqlNode: sqlWith.withList) {
        SqlWithItem sqlWithItem = (SqlWithItem) sqlNode;
        sqlWithItem.query.accept(this);
      }
      return sqlWith;
     } else if (sqlCall.getKind().equals(SqlKind.SELECT)) {
      SqlSelect sqlSelect = (SqlSelect) sqlCall;
      // no FROM clause return as it is
      if (sqlSelect.getFrom() == null) {
        return sqlSelect;
      }
      // There is an inner query, extract that and recursively expand that
      if (sqlSelect.getFrom().getKind().equals(SqlKind.SELECT)) {
        // this is mutating call so need to set From clause again in original query
        sqlSelect.getFrom().accept(this);
        return sqlSelect;
      }

      Artifact artifact = sqlSelect.getFrom().accept(new ExtractArtifact(artifactManager));
      // if the FROM clause refers a saved model, expand it
      if (artifact != null) {
        try {
          // 1. Get dimensions and measures list from the saved model
          // 2. For each column in the user query, check if it matches either a dimension or a measure in the model,
          //    it can refer to aliases as well
          // 3. If match is found expand them
          // 4. If no corresponding column found in model then throw an exception
          // 5. If some measures are present in the query then create a groupBy list of dimensions
          // 6. Prepare a select query using all these artifacts, it will be a group by query if any measures are present
          // 7. return the new query
          SqlCreateMetric sqlCreateMetric = calciteToolbox.parseModelingQuery(artifact.getPayload());
          Map<String, SqlNode> artifactMeasures = sqlCreateMetric.measuresMap;
          Map<String, SqlNode> artifactDimensions = sqlCreateMetric.dimensionsMap;

          SqlNodeList dimensions = new SqlNodeList(SqlParserPos.ZERO);
          SqlNodeList aggregates = new SqlNodeList(SqlParserPos.ZERO);

          for (SqlNode sqlNode : sqlSelect.getSelectList()) {
            String id = sqlNode.accept(new SimpleIdentifierExtractor(SimpleIdentifierExtractor.FROM.QUERY));
            if (id != null && id.equals("*")) {
              if (sqlSelect.getSelectList().size() > 1) {
                throw new RuntimeException("Cannot specify columns along with *");
              }
              dimensions.addAll(artifactDimensions.values());
              aggregates.addAll(artifactMeasures.values());
            } else if (id != null && artifactMeasures.containsKey(id)) {
              MeasureExpander measureExpander = new MeasureExpander(artifactMeasures.get(id));
              sqlNode = sqlNode.accept(measureExpander);
              aggregates.add(sqlNode);
            } else if (id != null && artifactDimensions.containsKey(id)) {
              DimensionExpander dimensionExpander = new DimensionExpander(artifactDimensions.get(id));
              sqlNode = sqlNode.accept(dimensionExpander);
              dimensions.add(sqlNode);
            } else {
              throw new SqlValidatorException(
                  String.format("Column [%s] not present in metrics view [%s]", sqlNode, artifact.getName()), null);
            }
          }

          SqlNodeList groupByList = new SqlNodeList(SqlParserPos.ZERO);
          if (!aggregates.isEmpty()) {
            for (int i = 1; i <= dimensions.size(); i++) {
              groupByList.add(SqlLiteral.createExactNumeric(i + "", SqlParserPos.ZERO));
            }
          }
          SqlNodeList selectList = new SqlNodeList(SqlParserPos.ZERO);
          selectList.addAll(dimensions);
          selectList.addAll(aggregates);
          sqlSelect.setSelectList(selectList);
          sqlSelect.setFrom(sqlCreateMetric.from);

          if (!groupByList.isEmpty()) {
            sqlSelect.setGroupBy(groupByList);
          }
          return sqlSelect;
        } catch (SqlParseException | SqlValidatorException e) {
          throw new RuntimeException(e);
        }
      }
    }
    // no reference to a model found or don't know what to do, return as it is
    return sqlCall;
  }
}
