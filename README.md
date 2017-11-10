# Time Resource

Implements a resource that reports new versions on a configured interval. The
interval can be arbitrarily long.

This resource is built to satisfy "trigger this build at least once every 5
minutes," not "trigger this build on the 10th hour of every Sunday." That
level of precision is better left to other tools.

## Source Configuration

* `interval`: *Optional.* The interval on which to report new versions. Valid
  values: `60s`, `90m`, `1h`.

* `skew`: *Optional.* A maximum bound of time by which to skew the trigger
  time. Valid values: `60s`, `90m`, `1h`. Must be used with `interval`.

* `location`: *Optional. Default `UTC`.* The
  [location](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) in
  which to interpret `start`, `stop`, and `days`.

  e.g.

  ```
  location: Africa/Abidjan
  ```

* `start` and `stop`: *Optional.* Only create new time versions between this
  time range. The supported formats for the times are: `3:04 PM`, `3PM`, `3
  PM`, `15:04`, and `1504`.

  e.g.

  ```
  start: 8:00 PM
  stop: 9:00 PM
  ```

  **Deprecation: an offset may be appended, e.g. `+0700` or `-0400`, but you
  should use `location` instead.**

* `days`: *Optional.* Run only on these day(s). Supported days are: `Sunday`,
  `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday` and `Saturday`.

  e.g.

  ```
  days: [Monday, Wednesday]
  ```

These can be combined to emit a new version on an interval during a particular
time period.

## Behavior

### `check`: Produce timestamps satisfying the interval.

Returns current version and new version only if it has been longer than `interval` since the
given version, or if there is no version given.


### `in`: Report the given time.

Fetches the given timestamp, writing the request's metadata to `input` in the
destination.

#### Parameters

*None.*


### `out`: Produce the current time.

Returns a version for the current timestamp. This can be used to record the
time within a build plan, e.g. after running some long-running task.

#### Parameters

*None.*


## Examples

### Periodic trigger

```yaml
resources:
- name: 5m
  type: time
  source: {interval: 5m}

jobs:
- name: something-every-5m
  plan:
  - get: 5m
    trigger: true
  - task: something
    config: # ...
```

### Periodic trigger with some skewing

```yaml
resources:
- name: 5-ish-min
  type: time
  source:
    interval: 5m
    skew: 1m

jobs:
- name: something-every-4-to-6m
  plan:
  - get: 5-ish-min
    trigger: true
  - task: something
    config: # ...
```

### Trigger once within time range

```yaml
resources:
- name: after-midnight
  type: time
  source:
    start: 12:00 AM
    stop: 1:00 AM
    location: Asia/Sakhalin

jobs:
- name: something-after-midnight
  plan:
  - get: after-midnight
    trigger: true
  - task: something
    config: # ...
```

### Trigger on an interval within time range

```yaml
resources:
- name: 5m-during-midnight-hour
  type: time
  source:
    interval: 5m
    start: 12:00 AM
    stop: 1:00 AM
    location: America/Bahia_Banderas

jobs:
- name: something-every-5m-during-midnight-hour
  plan:
  - get: 5m-during-midnight-hour
    trigger: true
  - task: something
    config: # ...
```

## Development

### Prerequisites

* golang is *required* - version 1.9.x is tested; earlier versions may also
  work.
* docker is *required* - version 17.06.x is tested; earlier versions may also
  work.
* godep is used for dependency management of the golang packages.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following command:

```sh
docker build -t time-resource .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
