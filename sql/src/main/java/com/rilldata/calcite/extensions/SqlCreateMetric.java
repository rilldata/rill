package com.rilldata.calcite.extensions;

import com.rilldata.calcite.generated.ParseException;
import com.rilldata.calcite.visitors.SimpleIdentifierExtractor;
import org.apache.calcite.sql.SqlCreate;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.SqlOperator;
import org.apache.calcite.sql.SqlSpecialOperator;
import org.apache.calcite.sql.SqlWriter;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.util.ImmutableNullableList;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * Example - North Star to support
 * <pre>
 * CREATE METRICS VIEW pageviews
 * DIMENSIONS
 * time GRANULARITIES(MINUTE, INTERVAL 15 MINUTE, HOUR, DAY),
 * CAST(user_created_at AS YEAR) AS cohort,
 * country COMMENT 'User''s country',
 * page TAG 'format' = 'url'
 * MEASURES
 * count(*) as views,
 * count(distinct user_id) as users,
 * avg(elapsed) as avg_elapsed
 * FROM clicks;
 * </pre>
 *
 * Currently, support something like -
 * <pre>
 * CREATE METRICS VIEW pageviews
 * DIMENSIONS
 * CAST(user_created_at AS YEAR) AS cohort,
 * country,
 * page
 * MEASURES
 * count(*) as views,
 * count(distinct user_id) as users,
 * avg(elapsed) as avg_elapsed
 * FROM clicks;
 * </pre>
 */
public class SqlCreateMetric extends SqlCreate
{
  public final SqlIdentifier name;
  public final SqlNodeList dimensions;
  public final SqlNodeList measures;
  public final SqlNode from;
  public final Map<String, SqlNode> dimensionsMap;
  public final Map<String, SqlNode> measuresMap;

  private static final SqlOperator OPERATOR =
      new SqlSpecialOperator("CREATE METRICS VIEW", SqlKind.OTHER_DDL);

  /**
   * Creates a SqlCreateMetric.
   */
  public SqlCreateMetric(SqlParserPos pos, SqlIdentifier name, SqlNodeList dimensions, SqlNodeList measures,
      SqlNode from
  ) throws ParseException
  {
    super(OPERATOR, pos, false, false);
    this.name = Objects.requireNonNull(name, "name");
    this.dimensions = Objects.requireNonNull(dimensions, "dimensions");
    this.measures = Objects.requireNonNull(measures, "measures");
    this.from = Objects.requireNonNull(from, "from");
    this.dimensionsMap = createMapFromList(dimensions);
    this.measuresMap = createMapFromList(measures);
  }

  @Override public List<SqlNode> getOperandList()
  {
    return ImmutableNullableList.of(name, dimensions, measures, from);
  }

  @Override public void unparse(SqlWriter writer, int leftPrec, int rightPrec)
  {
    writer.keyword("CREATE");
    writer.keyword("METRICS");
    writer.keyword("VIEW");
    name.unparse(writer, 0, 0);
    writer.newlineAndIndent();
    writer.keyword("DIMENSIONS");
    writer.newlineAndIndent();
    dimensions.unparse(writer, 0, 0);
    writer.newlineAndIndent();
    writer.keyword("MEASURES");
    writer.newlineAndIndent();
    measures.unparse(writer, 0, 0);
    writer.newlineAndIndent();
    writer.keyword("FROM");
    from.unparse(writer, 0, 0);
  }

  private static Map<String, SqlNode> createMapFromList(SqlNodeList sqlNodes) throws ParseException
  {
    SimpleIdentifierExtractor simpleIdentifierExtractor = new SimpleIdentifierExtractor(SimpleIdentifierExtractor.FROM.MODEL);
    // use LinkedHashMap to maintain insertion order which is same as how they appear in the model query
    Map<String, SqlNode> sqlNodeMap = new LinkedHashMap<>();
    for (SqlNode sqlNode : sqlNodes) {
      String name = sqlNode.accept(simpleIdentifierExtractor);
      if (name == null) {
        throw new ParseException(String.format("Cannot find identifier for node %s", sqlNode));
      }
      sqlNodeMap.put(name, sqlNode);
    }
    return sqlNodeMap;
  }
}
