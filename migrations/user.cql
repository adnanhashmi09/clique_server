CREATE KEYSPACE IF NOT EXISTS clique WITH REPLICATION = {'class': 'NetworkTopologyStrategy', 'replication_factor': 3};

USE clique;

CREATE TABLE IF NOT EXISTS users (
  id UUID ,
  username TEXT,
  email TEXT,
  name TEXT,
  password TEXT,
  rooms SET<UUID>,
  direct_msg_channels SET<UUID>,
  PRIMARY KEY ((id, username, email))
);


