# Changelog
Entries in this file should be limited to:
-  Any changes that introduce a deprecation in functionality, OR
-  Obscure side-effects that are not obviously apparent based on the JIRA associated with the changes.
Please avoid adding duplicate information across this changelog and JIRA/doc input pages.

## [NEXT RELEASE]
- Use migrated openshift-4-demo configuration.
- Add functionality to set the image registry, image tags, and helm chart versions in openshift-4-demo

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
  ref: https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
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
