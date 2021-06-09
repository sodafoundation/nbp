# Soda-syncer

Soda-syncer is an experimental feature for syncing the meta-data from different Platforms and adding features to SODA NBP Plugins.

Soda-syncer will be providing the following features :- 
 - [x] Consistent Snapshot to CSI Plugins
 - [ ] CSI Meta-data sync between Soda and K8s


*Note*: This is an experimental project and is not ready to be used in production.

### Consistent Snapshot Solution for CSI Plugins
Consistent Snapshot is considered as snapshots which are taken at regular intervals and pushed to cloud for DR Solutions.   
Consistent snapshot is provided by few CSI Plugin Drivers, but most of them lack this feature, Soda provides this solution for CSI Plugin Drivers which gives local PV. The local PV is backed up by soda-syncer at regular interval as configured and backed up to the cloud of your choice.
Soda leverages the Soda profile and [CSI Plug-N-Play](../csi-plug-n-play/) design to configure the snapshot policy and does the backup independently without platform support, currently this solution is available for K8s, and it doesn't require any operator/crd to add on this feature to existing CSI plugins.

![Consistent Snapshot Solution](static/assets/consistent-snapshot.png)