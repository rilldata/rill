package com.rilldata.calcite.extensions;

import org.apache.calcite.sql.SqlCreate;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlOperator;
import org.apache.calcite.sql.SqlSpecialOperator;
import org.apache.calcite.sql.SqlWriter;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.util.ImmutableNullableList;

import java.util.List;
import java.util.Map;

public class SqlCreateSource extends SqlCreate
{
  public final SqlIdentifier name;
  public final Map<String, String> properties;

  private static final SqlOperator OPERATOR =
      new SqlSpecialOperator("CREATE SOURCE", SqlKind.OTHER);

  public SqlCreateSource(SqlParserPos pos, SqlIdentifier name, Map<String, String> properties) {
    super(OPERATOR, pos, false, false);
    this.name = name;
    this.properties = properties;
  }

  @Override public void unparse(SqlWriter writer, int leftPrec, int rightPrec) {
    writer.keyword("CREATE SOURCE");
    name.unparse(writer, leftPrec, rightPrec);
    writer.keyword("WITH");
    writer.keyword("(");
    for (Map.Entry<String, String> entry : properties.entrySet()) {
      writer.keyword(entry.getKey());
      writer.keyword("=");
      writer.keyword(entry.getValue());
      writer.keyword(",");
    }
    writer.keyword(")");
  }

  @Override
  public List<SqlNode> getOperandList()
  {
    return ImmutableNullableList.of(name);
  }
}
