# Contributing

When contributing to this repository, please first discuss the change you wish
to make with the Maintainers by opening an issue with a description of the
problem you have. We may not accept every change, so do this first to make sure
you won't waste your time.

Before you do work on an open issue please say you do so in the issue. This
also avoids wasted effort.

## Setup

To run/debug tests in your IDE make sure the following environment variables for
your test runs:

| ENV VAR              | Value                                 |
|----------------------|---------------------------------------|
| `KUBEBUILDER_ASSETS` | `<absolute-path-to-repo>/testbin/bin` |

Required software is:

- Docker
- Go SDK

## Testing

The recommended way to test your changes is by both

- Testing things manually in a [**kind**](https://kind.sigs.k8s.io/) environment
- Adding tests to our existing end-to-end test suite

### Local Kind (Kubernetes in Docker) Environment

The `Makefile` includes targets to spin up a local kind environment and deploy
Heist, as well as all required dependencies in a fully automated way. Just run
the following command:

```shell
make kind/heist
```

Every time `kind/heist` is run a new Heist image is built and deployed in the
local kind cluster. This allows you to quickly test and verify that a change is
working as expected in an isolated environment.

### End-to-end testing

We use Ginkgo to run tests. Whenever you contribute a change it should
include the necessary tests to make sure that your change keeps working
in the future as expected.

For examples on how to write tests using Ginkgo, just look at our current
end-to-end tests.

In general, we use rely mostly on end-to-end tests. For that purpose our tests:

- Start a local Vault server in dev mode
- Start a local Kubernetes API server and etcd instance using
 [**setup-envtest**](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest).
  - At the moment we rely on an old version of `setup-envtest` that is just a
    bash script, we have not yet updated to the new go binary version.

This allows us to thoroughly test all Heist features in a lightweight and quick
way while at the same time ensuring that our test setup environment is as close
to a real production deployment as possible.

To run all end-to-end tests run the `test` target in the Makefile:

```shell
make test
```

There are other test targets prefixed with `test_` which only run a specific
part of the test suite. This is useful for a fast & efficient development workflow.

As of right now there is:

- `test_controllers`: Runs tests for all controllers
- `test_injector`: Runs tests for the agent injector
- `test_webhooks`: Runs Tests for the webhooks
- `test_agent`: Runs tests for the agent package
- `test_agentserver`: Runs tests for the agent server
- `test_vault`: Runs tests for our Vault API implementation

## GitHub Actions

We run our full test suite and validation for formatting and linting on every
commit. All those checks have to pass for a pull request to be merged. If you
add another end-to-end test Makefile target, you should also add the target to
the CI matrix in the [test workflow](./.github/workflows/test.yaml).

## Style Guidelines

### Code Style

Formatting and code style is enforced by our automated linting. To lint and
format your code run the `fix` target:

```yaml
make fix
```

Since we use the popular [golangci-lint](https://golangci-lint.run) for linting
and [gofumpt](https://github.com/mvdan/gofumpt) for formatting, there may be
integrations for your preferred IDE already.

In Jetbrains IDEs you can install the [Go Linter
plugin](https://plugins.jetbrains.com/plugin/12496-go-linter) and use File
Watchers to run gofumpt when you save the file.
Depending on the plugin, you may need to install those packages globally.

For MacOS those packages are available on Homebrew. Many GNU/Linux
Distributions have those packages available in the default package manager. You
can also install them using `go install`. For more information about that read
the documentation of those packages.

### Commit Style

#### Commits

Splitting commits is very important to us, so we can easily spot and revert a
commit which might have broken something.

A single commit should only affect a single aspect of the system and should not
include unrelated changes, like formatting. Commits should also be able to be
applied or reverted without resulting in a broken application. For more
information about this read [the wikipedia article about the atomic commit
convention](https://en.wikipedia.org/wiki/Atomic_commit#Atomic_commit_convention).

We also strongly prefer rebasing over merge commits. So pull requests must not
include merge commits.

#### Message

We use [**Conventional
Commits**](https://www.conventionalcommits.org/en/v1.0.0/) for our commit
messages. Any contributions are expected to also adhere to that commit format.

If you change anything related to Heist directly, we want to you use commit
scopes. Otherwise, you can omit the scope. Allowed scopes are:

- operator
- agent
- vault-api

## Licensing

Heist uses the [Apache 2.0 License](./LICENSE) to allow open contributions and
to give the users as many rights as possible.

It is important that you fully understand which rights you give up by
contributing.

### Developer Certificate of Origin

Heist requires the Developer Certificate of Origin (DCO) to be accepted on
every commit.

The DCO is here to make sure that we are actually allowed to merge your
contribution.

By adding a Signed-off-by statement (which you can do by using `git commit
-s`) in the contribution's commit message you agree to the DCO, which you can
find below or at [developercertificate.org](https://developercertificate.org/).

```txt
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

## Development Guidelines

### Designing new CRDs

The basic steps to create a new CRD are as follows:

1. Create go files under `pkg/apis/heist.youniqx.com/v1alpha1` to define the API
   for the new controller. Use existing controllers for reference

2. Implement the actual  logic in `pkg/controllers/<new_controller_name>`. Use existing
   controllers for reference.

3. Create End to End test suite in
   `pkg/controllers/e2e_test/<new_controller_name>`. Use existing tests for
   reference.

4. Create new target for e2e test in `Makefile` and add target to github actions
   matrix in `.github/workflows/test.yaml`

5. Execute `make generate`

#### Documentation

##### Descriptions

CRD Field Descriptions should be formatted as basic Markdown.
Use code-blocks to write examples.

##### Short Explanation and Examples

It is important to provide a short explanation what this new CRD is for, how it
works and some examples in <./docs/crds/>.

Also link this new document in our <./README.md> under *CRD Documentation*.

#### Validation

To validate the fields in the resource you can either use kubebuilder
annotations, custom webhook logic or both. If what you are doing requires only
basic validation that the kubebuilder annotations already offer, use them!

With custom validation logic you can also validate against Vault.

More Information:

- Kubebuilder CRD Validation Annotations: <https://book.kubebuilder.io/reference/markers/crd-validation.html>

#### Observability

We either use Logging to log to heist components, or use Kubernetes Events
& Conditions to attach status information to Kubernetes objects.

##### Logging

We use [logr](https://github.com/go-logr/logr) as the logging interface
throughout *Heist*.

Inside our controllers, we add the logging interface to the `Reconciler` struct
within the `controller.go` file and then initialize it with the logr
`WithValues` method, so we can add the controller name and Kubernetes object
information to the log.

Example: [pkg/controllers/vaultsecret/controller.go](https://github.com/youniqx/heist/blob/8514e41935ecef3794c4d73d35db58c1248f2a5e/pkg/controllers/vaultsecret/controller.go#L66)

You can then use the `Log` method from the
`Reconciler` struct throughout the controller.

##### Events

Events are meant to save history for human viewing. They provide information for
debugging, similar to log messages. Events are not intended for automation.

Will get documented into separate `kind: Event` Resources.

##### Conditions

Conditions are a standard status property which are used by other tools and
controllers to understand the state of this resource without needing to
implement resource-specific status details. Useful for automation, not human
consumption.

Will get documented as an array into the custom resource under `.status.condition`.

More Information:

- Kubernetes API Conventions: <https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties>

###### Predefined Values

We already have a set of ConditionReasons and ConditionTypes defined in
<./pkg/apis/heist.youniqx.com/v1alpha1/conditions.go>. Please use those
whenever possible, or define a new one in there. If a condition is specific to
this controller only and probably will never be used in another controller, it
is also fine to hardcode it locally in the controller.

#### End to End tests

You should also create an e2e test suite for your new controller, so we are
sure the CRD won't be broken by future changes.

##### Makefile target

Create a new target in <./Makefile> which executes your test suite with Ginkgo.
See the other test_controllers-* targets for examples.

##### Github Actions

You need to add your newly created target into the test matrix specified in
<./.github/workflows/test.yaml>, so it is executed by GH Actions in all PRs.

##### Flaky tests

Due to us requiring that every PR passes all test suites, please check with our
*flaky.sh* script that it does not fail due to a race condition or other random
issues. Please make sure the suite runs on your workstation at least **50** times
without a single failure.

The syntax for this script is `flaky.sh <makefile_target> <tries>`

Example:
`./utils/flaky.sh test_controllers_vaulttransitkey 50`
