# Change Log

## [Unreleased](https://github.com/kubedb/mysql/tree/HEAD)

[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.2...HEAD)

**Merged pull requests:**

- Concourse [\#41](https://github.com/kubedb/mysql/pull/41) ([tahsinrahman](https://github.com/tahsinrahman))
- Fixed kubeconfig plugin for Cloud Providers && Storage is required for MySQL [\#40](https://github.com/kubedb/mysql/pull/40) ([the-redback](https://github.com/the-redback))
- Remove lost+found directory before initializing mysql [\#39](https://github.com/kubedb/mysql/pull/39) ([the-redback](https://github.com/the-redback))
- Refactored E2E testing to support E2E testing with admission webhook in cloud [\#38](https://github.com/kubedb/mysql/pull/38) ([the-redback](https://github.com/the-redback))
- Skip delete requests for empty resources [\#37](https://github.com/kubedb/mysql/pull/37) ([the-redback](https://github.com/the-redback))
- Don't panic if admission options is nil [\#36](https://github.com/kubedb/mysql/pull/36) ([tamalsaha](https://github.com/tamalsaha))
- Disable admission controllers for webhook server [\#35](https://github.com/kubedb/mysql/pull/35) ([tamalsaha](https://github.com/tamalsaha))
- Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 [\#34](https://github.com/kubedb/mysql/pull/34) ([the-redback](https://github.com/the-redback))
- Update client-go to 7.0.0 [\#33](https://github.com/kubedb/mysql/pull/33) ([tamalsaha](https://github.com/tamalsaha))
- Added update script of docker for mysql-tools:8 [\#32](https://github.com/kubedb/mysql/pull/32) ([the-redback](https://github.com/the-redback))
- Added support of mysql:5.7 [\#31](https://github.com/kubedb/mysql/pull/31) ([the-redback](https://github.com/the-redback))
- Add support for one informer and N-eventHandler for snapshot, dromantdb and Job [\#30](https://github.com/kubedb/mysql/pull/30) ([the-redback](https://github.com/the-redback))
- Use metrics from kube apiserver [\#29](https://github.com/kubedb/mysql/pull/29) ([tamalsaha](https://github.com/tamalsaha))
- Bundle webhook server and Use SharedInformerFactory [\#28](https://github.com/kubedb/mysql/pull/28) ([the-redback](https://github.com/the-redback))
- Move MySQL AdmissionWebhook packages into MySQL repository [\#27](https://github.com/kubedb/mysql/pull/27) ([the-redback](https://github.com/the-redback))
- Use mysql:8.0.3 image as mysql:8.0 [\#26](https://github.com/kubedb/mysql/pull/26) ([the-redback](https://github.com/the-redback))
- Add travis yaml [\#25](https://github.com/kubedb/mysql/pull/25) ([tahsinrahman](https://github.com/tahsinrahman))

## [0.1.0-beta.2](https://github.com/kubedb/mysql/tree/0.1.0-beta.2) (2018-02-27)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.1...0.1.0-beta.2)

**Merged pull requests:**

- Migrating to apps/v1 [\#23](https://github.com/kubedb/mysql/pull/23) ([the-redback](https://github.com/the-redback))
- update validation [\#22](https://github.com/kubedb/mysql/pull/22) ([aerokite](https://github.com/aerokite))
-  Fix dormantDB matching: pass same type to Equal method [\#21](https://github.com/kubedb/mysql/pull/21) ([the-redback](https://github.com/the-redback))
- Use official code generator scripts [\#20](https://github.com/kubedb/mysql/pull/20) ([tamalsaha](https://github.com/tamalsaha))
-  Fixed dormantdb matching & Raised throttling time & Fixed MySQL version Checking [\#19](https://github.com/kubedb/mysql/pull/19) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.1](https://github.com/kubedb/mysql/tree/0.1.0-beta.1) (2018-01-29)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.0...0.1.0-beta.1)

**Merged pull requests:**

- converted to k8s 1.9 & Improved InitSpec in DormantDB & Added support for Job watcher & Improved Tests [\#17](https://github.com/kubedb/mysql/pull/17) ([the-redback](https://github.com/the-redback))
- Fixed logger, analytics and Removed rbac stuff [\#16](https://github.com/kubedb/mysql/pull/16) ([the-redback](https://github.com/the-redback))
- Add rbac stuffs for mysql-exporter [\#15](https://github.com/kubedb/mysql/pull/15) ([the-redback](https://github.com/the-redback))
-  Review Mysql docker images and Fixed monitring [\#14](https://github.com/kubedb/mysql/pull/14) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.0](https://github.com/kubedb/mysql/tree/0.1.0-beta.0) (2018-01-07)
**Merged pull requests:**

- Rename ms-operator to my-operator [\#13](https://github.com/kubedb/mysql/pull/13) ([tamalsaha](https://github.com/tamalsaha))
- Fix Analytics and pass client-id as ENV to Snapshot Job [\#12](https://github.com/kubedb/mysql/pull/12) ([the-redback](https://github.com/the-redback))
- update docker image validation [\#11](https://github.com/kubedb/mysql/pull/11) ([the-redback](https://github.com/the-redback))
- Add docker-registry and WorkQueue  [\#10](https://github.com/kubedb/mysql/pull/10) ([the-redback](https://github.com/the-redback))
- Set client id for analytics [\#9](https://github.com/kubedb/mysql/pull/9) ([tamalsaha](https://github.com/tamalsaha))
- Fix CRD Registration [\#8](https://github.com/kubedb/mysql/pull/8) ([the-redback](https://github.com/the-redback))
- Update pkg paths to kubedb org [\#7](https://github.com/kubedb/mysql/pull/7) ([tamalsaha](https://github.com/tamalsaha))
- Assign default Prometheus Monitoring Port [\#6](https://github.com/kubedb/mysql/pull/6) ([the-redback](https://github.com/the-redback))
- mysql-util docker image [\#5](https://github.com/kubedb/mysql/pull/5) ([the-redback](https://github.com/the-redback))
- Add Snapshot Backup, Restore and Backup-Scheduler [\#4](https://github.com/kubedb/mysql/pull/4) ([the-redback](https://github.com/the-redback))
- Update ./hack folder [\#3](https://github.com/kubedb/mysql/pull/3) ([tamalsaha](https://github.com/tamalsaha))
- Mysql db - Inititalizing  [\#2](https://github.com/kubedb/mysql/pull/2) ([the-redback](https://github.com/the-redback))
- Add skeleton for mysql [\#1](https://github.com/kubedb/mysql/pull/1) ([aerokite](https://github.com/aerokite))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*