# Time Resource

Implements a resource that reports new versions on a configured interval. The
interval can be arbitrarily long.

This resource is built to satisfy "trigger this build at least once every 5
minutes," not "trigger this build on the 10th hour of every Sunday." That
level of precision is better left to other tools.

## Source Configuration

* `interval`: *Optional.* The interval on which to report new versions. Valid
  units are: “ns”, “us” (or “µs”), “ms”, “s”, “m”, “h”. Examples: `60s`, `90m`,
  `1h30m`. If not specified, this resource will generate exactly 1 new version
  per calendar day on each of the valid `days`.

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

  **Note: YAML parsers like PyYAML may parse time values in the 24h format as integers, not strings (e.g. `19:00` is parsed as `1140`). If you pre-process your pipeline configuration with such a parser this might trigger a marshaling error. In that case you can quote your `start` and `stop` values, so they will be correctly treated as string.**

* `days`: *Optional.* Limit the creation of new time versions to the specified
  day(s). Supported days are: `Sunday`, `Monday`, `Tuesday`, `Wednesday`,
  `Thursday`, `Friday` and `Saturday`.

  e.g.

  ```
  days: [Monday, Wednesday]
  ```

  These can be combined to emit a new version on an interval during a particular
  time period.

* `initial_version`: *Optional.* When using `start` and `stop` as a trigger for
  a job, you will be unable to run the job manually until it goes into the
  configured time range for the first time (manual runs will work once the `time`
  resource has produced it's first version).

  To get around this issue, there are two approaches:
     * Use `initial_version: true`, which will produce a new version that is
       set to the current time, if `check` runs and there isn't already a version
       specified. **NOTE: This has a downside that if used with `trigger: true`, it will
       kick off the correlating job when the pipeline is first created, even
       outside of the specified window**.
     * Alternatively, once you push a pipeline that utilizes `start` and `stop`, run the
       following fly command to run the resource check from a previous point
       in time (see [this issue](https://github.com/concourse/time-resource/issues/24#issuecomment-689422764)
       for 6.x.x+ or [this issue](https://github.com/concourse/time-resource/issues/11#issuecomment-562385742)
       for older Concourse versions).

       ```
       fly -t <your target> \
         check-resource --resource <pipeline>/<your resource>
         --from "time:2000-01-01T00:00:00Z" # the important part
       ```

       This has the benefit that it shouldn't trigger that initial job run, but
       will still allow you to manually run the job if needed.

  e.g.

  ```
  initial_version: true
  ```

## Behavior

### `check`: Produce timestamps satisfying the interval.

Returns current version and new version only if it has been longer than `interval` since the
given version, or if there is no version given.


### `in`: Report the given time.

Fetches the given timestamp. Creates two files:
1. `input` which contains the request provided by Concourse
1. `timestamp` which contains the fetched version in the following format: `2006-01-02 15:04:05.999999999 -0700 MST`

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
docker build -t time-resource --target tests -f dockerfiles/alpine/Dockerfile .
docker build -t time-resource --target tests -f dockerfiles/ubuntu/Dockerfile .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
