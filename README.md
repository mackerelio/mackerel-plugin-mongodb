mackerel-plugin-mongodb
=====================

MongoDB custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-mongodb [-host=<host>] [-port=<port>] [-username=<username>] [-password=<password>] [-tempfile=<tempfile>] [-source=<authenticationDatabase>]
```

```shell
mackerel-plugin-mongodb [-url=<mongodb://.....>] [-tempfile=<tempfile>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.mongodb]
command = "/path/to/mackerel-plugin-mongodb"
```

## Add Role

newer mongodb requre `clusterMonitor` role when executed `db.serverStatus()` command.

so add role `clusterMonitor` to reporter.

```
db.grantRolesToUser(
  "user_id",
  [
  { role: "clusterMonitor", db:"admin"}
  ]
 );
 ```

see https://dba.stackexchange.com/questions/121832/db-serverstatus-got-not-authorized-on-admin-to-execute-command

## Supported MongoDB versions

* `v1.1.0` mongodb 3.6, 4.0, 4.2, 4.4, 5.0, 6.0 or later
* `v1.0.0` mongodb 2.2, 2.6, 3.0, 3.2, 3.4, 3.6, 4.0, 4.2

## Supported Operating Systems

* Linux
