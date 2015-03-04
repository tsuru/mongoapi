mongoapi
========

[![Build Status](https://secure.travis-ci.org/tsuru/mongoapi.png?branch=master)](http://travis-ci.org/tsuru/mongoapi)

MongoDB service API for [tsuru PaaS](https://tsuru.io).

##Installation and configuration

This API is ready for being deployed as a tsuru application. It depends on the
following environment variables:

* ``MONGODB_URI``: The connection string of the MongoDB server that the API
  should use. _Default value:_ 127.0.0.1:27017;
* ``MONGODB_PUBLIC_URI``: URI in the format <host>:<port> used to access the
  MongoDB server externally. _Default value:_ the value of ``MONGODB_URI``;
* ``MONGODB_REPLICA_SET``: name of the replica set in use. It's optional, when
  ommited, the API won't use a replica set;
* ``MONGOAPI_DBNAME``: name of the database in use to store API metadata.
