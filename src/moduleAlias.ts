import moduleAlias from "module-alias";
moduleAlias.addAliases({
    "$lib": __dirname + "/lib",
    "$common": __dirname + "/common",
    "$cli": __dirname + "/cli",
    "$server": __dirname + "/server"
});
