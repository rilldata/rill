CREATE EXTENSION "uuid-ossp";

CREATE TABLE catalog (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE UNIQUE INDEX catalog_name_idx ON catalog (name);
