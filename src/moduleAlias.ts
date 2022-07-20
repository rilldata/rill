import moduleAlias from "module-alias";
import { __dirname } from "./common/utils/commonJsPaths";

moduleAlias.addAliases({
  $lib: __dirname + "/lib",
  $common: __dirname + "/common",
  $cli: __dirname + "/cli",
  $server: __dirname + "/server",
});
