# Configuration

The Coldwire-server is configured via a `JSON` file

There are 2 types of database storages in the server: 
1. User storage
2. Data storage

Each storage serves different purpose, the `User storage` is where to store the user IDs, and their respective authentication public-keys. It also stores the temporary cryptographic signing `challenges` given to users in order to authenticate.

While the `Data storage` stores actual data to be relayed between different users and or servers.


By default, the configuration file uses "`internal`" for both, which uses SQLite3 under the hood. 
This is usually fine for small-to-medium userbases, but if you expect a lot of users, and or better DDoS handling, you may want to switch to other storage mechanisms. 

Currently, we support the following storage options for `User storage`:
- Internal (SQLite3)
- SQL (MySQL, MariaDB, etc.)


And we support the following storage options for `Data storage`:
- Internal (SQLite3)
- SQL (MySQL, MariaDB, etc.)
- Redis


If you are facing performance problems, we highly recommend using SQL for `User Storage` and either `SQL` or `Redis` for `Data storage`.

