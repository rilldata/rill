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

/**
 * Finds the dependencies of a database entity, ie can find what tables a view depends on.
 * For example, given the following view:
 * create view C as select * from B join A on B.a = A.a
 * The dependencies of C are B and A.
 * Transient dependencies are not included, for example for a dependency tree C->B->A, C depends on A (B is transient).
 */
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
    if (node instanceof SqlCreateView create) {
      visit(create.query);
    } else if (node instanceof SqlCreateTable create) {
      visit(create.query);
    }
    return dependencies;
  }
}
