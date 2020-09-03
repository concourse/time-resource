# Time Resource

Implements a resource that reports new versions on a configured interval. The
interval can be arbitrarily long.

This resource is built to satisfy "trigger this build at least once every 5
minutes," not "trigger this build on the 10th hour of every Sunday." That
level of precision is better left to other tools.

## Source Configuration

* `interval`: *Optional.* The interval on which to report new versions. Valid
  examples: `60s`, `90m`, `1h`. If not specified, this resource will generate
  exactly 1 new version per calendar day on each of the valid `days`.

* `location`: *Optional. Default `UTC`.* The
  [location](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) in
  which to interpret `start`, `stop`, and `days`.

  e.g.

  ```
  location: Africa/Abidjan
  ```

* `start` and `stop`: *Optional.* Limit the creation of new versions to times
  on/after `start` and before `stop`. The supported formats for the times are:
  `3:04 PM`, `3PM`, `3PM`, `15:04`, and `1504`. If a `start` is specified, a
  `stop` must also be specified, and vice versa. If neither value is specified,
  both values will default to `00:00` and this resource can generate a new
  version (based on `interval`) at any time of day.

  e.g.

  ```
  start: 8:00 PM
  stop: 9:00 PM
  ```

  **Deprecation: an offset may be appended, e.g. `+0700` or `-0400`, but you
  should use `location` instead.**

  To explicitly represent a full calendar day, set `start` and `stop` to
  the same value.

  e.g.

  ```
  start: 6:00 AM
  stop: 6:00 AM
  ```

* `days`: *Optional.* Limit the creation of new time versions to the specified
  day(s). Supported days are: `Sunday`, `Monday`, `Tuesday`, `Wednesday`,
  `Thursday`, `Friday` and `Saturday`.

  e.g.

  ```
  days: [Monday, Wednesday]
  ```

These can be combined to emit a new version on an interval during a particular
time period.

## Behavior

### `check`: Produce timestamps satisfying the interval.

Returns a list of any new versions for which the `interval` has elapsed since the
previous version. If no version is given, returns the most recent timestamp that
satisfies the configuration.


### `in`: Report the given time.

Returns a version for most recent timestamp that satisfies the configuration which is
at/before the given timestamp. The request's metadata is written to `input` in the
destination.

#### Parameters

*None.*


### `out`: Produce the current time.

Returns a version for most recent timestamp that satisfies the configuration.

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
* go mod is used for dependency management of the golang packages.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following commands for both `alpine` and `ubuntu` images:

```sh
docker build -t time-resource -f dockerfiles/alpine/Dockerfile .
docker build -t time-resource -f dockerfiles/ubuntu/Dockerfile .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
