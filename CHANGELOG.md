# Changelog

This CHANGELOG is abandoned in favor of auto-generated release notes.
Please see the release description for change information.

## [0.8.14]

- ROX-17343: Setup metric collection and alerting for infra service.

## [0.8.13]

- ROX-22142: wait for IBM ROKS nodes to be ready during provisioning.
- Fix GKE cluster destruction with retries and disabled auto-upgrades and repairs.
- Create GCP workload identity pools for `openshift-4` clusters. They can be used to setup short-lived token integrations.

## [0.8.12]

- Add IBM ROKS flavor. Note that there are two known issues (ROX-22142, ROX-22143).

## [0.8.11]

- Bump default demo versions to 4.3.4.
- Bump demo automation flavors images to 0.9.13.

- Add an optional parameter to the in openshift-4 flavor to keep failed clusters around (infra will still delete the cluster)

## [0.8.10]

- Adds a fix for `demo` and `qa-demo` flavors: Region parameters should be optional on cluster creation, defaults are set in UI and the flavor itself.

## [0.8.9]

- Add region configuration to more flavors. This is now supported by all except ARO and AKS.
- Add flavor aliases, e.g. `ocp-4` for `openshift-4`. This may be used for `infractl flavor get` and `infractl create` commands.
- Bump default demo versions to 4.3.2.

## [0.8.8]

- Add cluster parameters to the `infractl get <cluster-id> --json` output to help debugging.
- Bump default demo versions to 4.3.0.
- Add markdown for parameter help formatting.
- Fix an issue with PWA.

## [0.8.7]

- Add support for credentialsMode to openshift-4* flavors and default to
  Passthrough to mimic CI.

## [0.8.6]

- Chore: Bump demo versions to 4.2.2

## [0.8.5]

- Fix: Openshift 3.11 provisioning is now based on quay.io/openshift installers (ROX-20327)

## [0.8.4]

- Fix: Scrolling on the cluster info page works again

## [0.8.3]

- Deploying infra-server with Helm and GCP Secret Manager

## [0.8.2]

- Hackathon '23:
  - Adding all clusters to url and drag/drop window split screen
  - ROX-17295: trusted certificates for openshift-4
- Fix: CLI and UI now consistently enforce restrictions on the cluster name format and length
- Change: ROX-19218, etc: Move GCP based OCP flavors in infra to a RH project
- Change: Use more consistent artifact naming for openshift clusters
- Chore: Misc tailwind -> patternfly
- Chore: Bump demo versions to 4.2.1

## [0.8.0]

- Switch GKE based flavors (gke-default, demo, qa-demo) to use a RH project (ROX-17123,ROX-19217)
- CLI: Add client-side cluster name validation
- Bump demo flavors to 4.2.0

## [0.7.11]

- Make the domain for GKE based demos configurable.
- Revert the change that reduced the master node count of openshift-4 and openshift-4-demo flavors from
  3 to 1 (default is now 3 again).

## [0.7.10]

- Fix for openshift-4-demo 4.2+ installs
- Reduce the master node count of openshift-4 and openshift-4-demo flavors from
  3 to 1
- Bump demo flavors to 4.1.3
- More pattern fly UI components

## [0.7.9]

- Upgrade Go version to 1.20
- Record cluster creation and deletions in BigQuery.
- Bump demo flavors to 4.1.2

## [0.7.8]

- Bump demo flavors to 4.1.1

## [0.7.7]

- Use latest openshift-4-demo to work with 4.1 rcs.

## [0.7.6]

- Bump demo flavors to 4.0.2

## [0.7.5]

- Fix ROSA flavor by pinning CLI versions.

## [0.7.4]

- Bump demo flavors to 4.0.0.

## [0.7.3]

- Update Go dependencies to close Dependabot alerts.

## [0.7.2]

- Configuration of all demo contents is fixed in the Openshift 4 Demo flavor.

## [0.7.1]

- ROX-15237: The Openshift 4 Demo flavor now supports testing of unreleased versions.
- Bump demo flavors to 3.74.2
- Add central-db-image parameter to qa-demo flavor
- --endpoint flag will now accept also URLs like <https://infra.rox.systems>, and addresses without a port like infra.rox.systems.

## [0.7.0]

- ROX-14317: New INFRA_TOKEN format includes a validity period. All existing tokens will be invalid and all users (humans and robots) need to regenerate their tokens.

## [0.6.6]

- Bump demo and qa-demo to 3.74.1

## [0.6.5]

- ROX-1251: Restore SSH for OpenShift-4 clusters
- ROX-15126: Add Openshift 4 flavor for testing performance and scaling
- Fix EKS by using the latest EKS automation-flavor.

## [0.6.4]

- Fix Slack notifications after migration to Internal Red Hat.

## [0.6.3]

- Use migrated openshift-4-demo configuration.
- Add functionality to set the image registry, image tags, and helm chart versions in openshift-4-demo
- Add a FIPS toggle to the openshift-4 flavor
- Improve ROSA cluster create logs
- Improve help message for openshift-version in the openshift-4 flavor to explain OCP dev previews
- Bump demo and qa-demo to 3.74.0 (openshift-4-demo uses latest by default)

## [0.6.2]

- Bump demo flavors to the latest 3.73.2 release.

## [0.6.1]

- Notice CLI users only about minor or major version mismatch (previously also for patch versions).
- Bump demo flavors to the latest 3.73.0 release.

## [0.6.0]

- Get a QA demo with just `infractl create qa-demo`. When `infractl create
  qa-demo` is run without specifying a NAME in a stackrox/stackrox repo context
  a NAME will be generated based on user initials and the most recent commit.
  The most recent commit will also be used to set main-image.
- Default names for other contexts. When `infractl create` is run without
  specifying a NAME one will be generated based on user initials, date and a
  counter to ensure uniqueness e.g. jb-10-31-1.
- Longer log retention. For usability and troubleshooting, logs are now kept for
  30 days (versus the previous 1 day.)
- `infractl status` command manages maintenance status of an infra deployment to
  influence Continuous Deployment.

## [0.5.3]

- Switch to containerd for GKE runtime to support k8s v1.23. Affects
  gke-default, demo, qa-demo.

## [0.5.2]

- Remove the --license arg for newer ACS installs.

## [0.5.1]

- gke-default: Use the gke-gcloud-auth-plugin for kubeconfig artifact. (#719)
  ref: <https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke>
  This change may require users to update their gcloud SDK and install the auth plugin.
  For MacOS users: `gcloud components update && gcloud components install gke-gcloud-auth-plugin`.
- Bump demo images to latest 3.72.1 release (#723)
- RS-576: Add default Prometheus metrics (#721)

## [0.5.0]

### Removed Features

### Deprecated Features

### Technical Changes

- Breaking change: `infractl get` output in JSON format now contains a string for the status instead of an enum.
- Artifacts produced by GKE and AKS clusters now come with set file permissions.
- Migrate from Circle CI to Github Actions for continuous integration.
- Upgrade to Go 1.17.
