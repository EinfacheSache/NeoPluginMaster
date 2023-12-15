# NeoPluginMaster

The NeoPluginMaster is a program that serves as a master for the NeoProtect plugin, collecting specific data such as plugin version, online players, and server version.

## Functionality

The NeoPluginMaster program is capable of collecting various metrics from the NeoProtect plugin. These metrics include:

- **PlayerAmount:** Number of online players on the server.
- **ServerAmount:** General server metrics.
- **ManageServer:** Metrics related to server management.
- **PluginVersion:** Current version of the NeoProtect plugin.
- **ServerVersion:** Version of the server software in use.
- **ServerName:** Name of the server.
- **VersionStatus:** Status metrics related to the versions.
- **VersionError:** Error metrics related to the versions.
- **UpdateSetting:** Metrics concerning update settings.
- **NeoProtectPlan:** Information about the NeoProtect plan.
- **JavaVersion:** Version of Java installed.
- **OsName:** Operating system name.
- **OsArch:** Operating system architecture.
- **OsVersion:** Operating system version.
- **CoreCount:** Number of CPU cores.
- **OnlineMode:** Server's online mode status.
- **ProxyProtocol:** Proxy protocol status.

The ServerStats metric is a collection of all data from a plugin client in a Prometheus output format. The plugin sends this data to Prometheus in a processed version for proper visualization. Moreover, inactive servers that are offline are removed from the data to maintain accuracy.
