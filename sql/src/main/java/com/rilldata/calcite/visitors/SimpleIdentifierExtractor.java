package com.rilldata.calcite.visitors;

import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.util.SqlBasicVisitor;

/**
 * This is used to extract identifier used as keys to store/retrieve dimension and measure definitions.
 * This is used both in modeling and user queries.
 *
 * <pre>
 *     For example - MODEL-
 *                     CREATE METRICS VIEW ...
 *                     MEASURES ...
 *                     avg(time) AS avg_elapsed
 *                     ...
 *     For above, <b>avg(time) AS avg_elapsed</b>, it will return <b>avg_elapsed</b>
 *
 *                     QUERY - SELECT ... avg_elapsed AS RANDOM_NAME ... FROM <METRICS_VIEW>
 *      For above query, it will return actual key of <b>avg_elapsed</b>
 *
 * Currently, its does not support compound identifiers so for things like
 *    SELECT p.avg_elapsed FROM PAGEVIEWS P
 * </pre>
 * where PAGEVIEWS is a model will not work and null will be returned.
 *
 * */
public class SimpleIdentifierExtractor extends SqlBasicVisitor<String>
{
  // Represents whether extraction is happening for a modeling query or a user query
  public enum FROM
  {
    MODEL,
    QUERY
  }

  private final FROM from;

  public SimpleIdentifierExtractor(FROM from)
  {
    this.from = from;
  }

  @Override public String visit(SqlCall call)
  {
    // Operand index will be last in case of model
    // but first in case of actual query
    // For example -  CREATE METRICS VIEW ...
    //                MEASURES ...
    //                avg(time) AS avg_elapsed
    //                ...
    //
    //                QUERY - SELECT ... avg_elapsed AS RANDOM_NAME ... FROM <METRICS_VIEW>
    if (from.equals(FROM.QUERY)) {
      return call.operand(0).accept(this);
    } else {
      return call.operand(call.operandCount() - 1).accept(this);
    }
  }

  @Override public String visit(SqlIdentifier id)
  {
    return id.isSimple() ? id.getSimple().toLowerCase() : (id.isStar() ? "*" : null);
  }
}
