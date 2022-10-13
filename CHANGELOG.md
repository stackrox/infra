# Changelog
Entries in this file should be limited to:
-  Any changes that introduce a deprecation in functionality, OR
-  Obscure side-effects that are not obviously apparent based on the JIRA associated with the changes.
Please avoid adding duplicate information across this changelog and JIRA/doc input pages.

## [NEXT RELEASE]

## [0.5.0]

### Removed Features
### Deprecated Features
### Technical Changes
- Breaking change: `infractl get` output in JSON format now contains a string for the status instead of an enum.
- Artifacts produced by GKE and AKS clusters now come with set file permissions.
- Migrate from Circle CI to Github Actions for continuous integration.
- Upgrade to Go 1.17.
