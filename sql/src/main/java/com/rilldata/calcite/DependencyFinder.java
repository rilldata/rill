package com.rilldata.calcite;

import org.apache.calcite.sql.SqlBasicCall;
import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlJoin;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlSelect;
import org.apache.calcite.sql.ddl.SqlCreateTable;
import org.apache.calcite.sql.ddl.SqlCreateView;
import org.apache.calcite.sql.fun.SqlStdOperatorTable;
import org.apache.calcite.sql.util.SqlBasicVisitor;

import java.util.ArrayList;
import java.util.List;

public class DependencyFinder extends SqlBasicVisitor<List<String>>
{
  List<String> dependencies = new ArrayList<>();

  public void visit(SqlNode from1)
  {
    if (from1 instanceof SqlIdentifier from11) {
      dependencies.add(from11.toString());
    } else if (from1 instanceof SqlBasicCall from11 && from11.getOperator() == SqlStdOperatorTable.AS) {
      visit(from11.getOperandList().get(0));
    } else if (from1 instanceof SqlJoin from11) {
      visit((from11).getLeft());
      visit((from11).getRight());
    } else if (from1 instanceof SqlSelect from11) {
      if (from11.getFrom() != null) {
        visit(from11.getFrom());
      }
    } else if (from1 instanceof SqlBasicCall from11 && from11.getOperator() == SqlStdOperatorTable.UNION_ALL) {
      for (SqlNode node : from11.getOperandList()) {
        visit(node);
      }
    }
  }

  @Override
  public List<String> visit(SqlCall node)
  {
//      for (SqlNode node : s) {
    if (node instanceof SqlCreateView create) {
      visit(create.query);
    } else if (node instanceof SqlCreateTable create) {
      visit(create.query);
    }
//      }
    return dependencies;
  }
}
