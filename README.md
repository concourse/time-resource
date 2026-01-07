# Time Resource

Implements a resource that reports new versions on a configured interval or cron schedule. The time intervals can be
arbitrarily long.

<a href="https://ci.concourse-ci.org/teams/main/pipelines/resource/jobs/build?vars.type=%22time%22">
  <img src="https://ci.concourse-ci.org/api/v1/teams/main/pipelines/resource/jobs/build/badge?vars.type=%22time%22" alt="Build Status">
</a>

This resource is built to satisfy needs like "trigger this build at least once every 5 minutes" or "trigger this build
at specific times using cron expressions." For simple interval-based triggering, the interval configuration is simpler
to use. For more complex scheduling, the cron configuration provides greater flexibility.

## Source Configuration

### Interval-based Configuration

* `interval`: *Optional.* The interval on which to report new versions. Valid
  units are: "s", "m", "h". Examples: `60s`, `90m`, `1h30m`. If not specified, this resource will
  generate exactly 1 new version per calendar day on each of the valid `days`.

* `location`: *Optional. Default `UTC`.* The
  [location](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) in
  which to interpret `start`, `stop`, and `days`.

  e.g.

  ```
  location: Africa/Abidjan
  ```

* `start` and `stop`: *Optional.* Limit the creation of new versions to times
  on/after `start` and before `stop`. The supported formats for the times are:
  `3:04 PM`, `3PM`, `3 PM`, `15:04`, and `1504`. If a `start` is specified, a
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

### Cron-based Configuration

* `cron`: *Optional.* A cron expression that defines when new versions should be created. Standard cron format is
  supported with 5 fields (minute, hour, day of month, month, day of week).

  **Important:** The 6-field format (including seconds) is not supported. You must use the standard 5-field format that
  only specifies minute precision.

  e.g.

  ```
  cron: "0 * * * *"  # Every hour at minute 0
  ```

**Tags:** Shorthand aliases for common schedules:

| Tag | Expression | Schedule |
|-----|------------|----------|
| `@yearly` / `@annually` | `0 0 1 1 *` | Midnight, Jan 1st |
| `@monthly` | `0 0 1 * *` | Midnight, 1st of month |
| `@weekly` | `0 0 * * 0` | Midnight, Sunday |
| `@daily` | `0 0 * * *` | Midnight |
| `@hourly` | `0 * * * *` | Start of every hour |
| `@30minutes` | `0,30 * * * *` | Every 30 minutes (:00, :30) |
| `@15minutes` | `*/15 * * * *` | Every 15 minutes |
| `@10minutes` | `*/10 * * * *` | Every 10 minutes |
| `@5minutes` | `*/5 * * * *` | Every 5 minutes |

e.g.
  ```
  cron: "@daily"  # Run once a day at midnight
  ```

**Modifiers:** Special modifiers for complex scheduling:

| Field | Modifier | Example | Description |
|-------|----------|---------|-------------|
| Day of Month | `L` | `0 2 L * *` | Last day of month (e.g., 28th/29th/30th/31st) |
| | `W` | `0 1 15W * *` | Nearest weekday to date (if 15th is Sat, triggers Fri 14th) |
| Day of Week | `L` | `0 3 * * 5L` | Last occurrence in month (5L = last Friday) |
| | `#` | `0 5 * * 1#2` | Nth occurrence in month (1#2 = second Monday) |

**Note: You cannot use `cron` together with `interval`, `start`, `stop`, or `days`. Use either the cron-based or
interval-based configuration.**

* `location`: *Optional. Default `UTC`.* When used with `cron`, the cron schedule is evaluated in this timezone.  
  See interval-based configuration above for format details.

### Common Configuration Options

* `initial_version`: *Optional.* When using `start` and `stop` or `cron` as a trigger for
  a job, you will be unable to run the job manually until it reaches the
  configured time range or cron schedule for the first time (manual runs will work once the `time`
  resource has produced its first version).

  To get around this issue, there are two approaches:
    * Use `initial_version: true`, which will produce a new version that is
      set to the current time, if `check` runs and there isn't already a version
      specified. **NOTE: This has a downside that if used with `trigger: true`, it will
      kick off the correlating job when the pipeline is first created, even
      outside of the specified window**.
    * Alternatively, once you push a pipeline that utilizes time-based constraints, run the
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
* `start_after`: *Optional.* Specifies the earliest datetime from which new time-based versions can be created.

  Supported formats are `2006-01-02 15:04:05`, `2006-01-02T15:04:05`, `2006-01-02T15:04`, `2006-01-02T15`, `2006-01-02`.

  Behavior:
    - If the `start_after` datetime is specified and is in the future, it will determine when the first version is created.
    - If the `start_after` datetime is in the past, the resource will continue to generate versions based on the other configuration parameters.
    - When `initial_version` is set to true, the first version will be created based on the current time. Subsequent versions will only be generated if they fall after the `start_after` datetime.
    - If a `location` is provided, the `start_after` datetime will be interpreted in the context of the specified timezone, rather than in UTC.

  e.g.

  ```
  start_after: 2023-10-01T00:00:00
  ```

### Differences Between `interval` and `cron`

There is a difference between `interval` and `cron` when trying to create similar schedules. `interval` will trigger regardless of calendar boundaries, while `cron` will trigger strictly following calendar boundaries. Let's look at an example.

If we want something to run "every 2 days" you can do that in these two ways:

* `interval: 48h` or
* `cron: "0 0 */2 * *"`

When these configurations trigger is very different.

The `interval` configuration will trigger every 48 hours based on when the last trigger ran.

The `cron` configuration will trigger every 2 calendar days at midnight. Cron also calculates "every 2 days" to be the 1st of each month and then every 2 days from then. So this cron schedule will trigger on the 1st, 3rd, 5th, 7th, etc. of every month. This also means if you're in a month with a 31st day, the resource will emit a version on the 31st and then again on the 1st of the next month, resulting in a trigger two days in a row.

A similar convention is followed with minutes and hours. When trying to schedule cron intervals like "every x minute/hour", cron will actually trigger "every x minute/hour of the hour/day". For example:

* `*/5 * * * *` "Every 5 minutes" is actually "every 5th minute of the hour" (00, 05, 10, 15, etc.)
* `0 */6 * * *` "Every 6 hours" is actually "every 6th hour of the day" (00, 06, 12, 18)

**Recommendation:** If you want true elapsed-time intervals (e.g., "every 48 hours from the last run"), use `interval`. If you want calendar-aligned schedules (e.g., "at midnight on specific days"), use `cron`.

### Cron Diagnostic Output

When a cron-triggered version is emitted, the resource logs a human-readable explanation to stderr:

```
cron: emitting version at 2025-01-07T00:00:00Z (previous: 2025-01-05T00:00:00Z)
  triggers every 2 days from 1st of month, at 00:00; note: 31st then 1st = back-to-back triggers
```

This includes warnings for common cron pitfalls:

| Condition | Warning |
|-----------|---------|
| Day step lands on 31st (e.g., `*/2`, `*/3`, `*/5`) | `note: 31st then 1st = back-to-back triggers` |
| Day 31 specified | `note: only triggers in months with 31 days (Jan, Mar, May, Jul, Aug, Oct, Dec)` |
| Day 30 specified | `note: skips February` |
| Day 29 specified | `note: only triggers in leap years for February` |
| Both day-of-month and day-of-week set | `note: day-of-month AND day-of-week uses OR logic, not AND (triggers on EITHER match)` |
| Hour 1-3 specified | `note: may skip or double-trigger during DST transitions` |

## Behavior

### `check`: Produce timestamps satisfying the interval or cron schedule.

Returns current version and new version only if it has been longer than `interval` since the
given version, or if the time matches the specified cron expression, or if there is no version given.

### `in`: Report the given time.

Fetches the given timestamp. Creates three files:
1. `input` which contains the request provided by Concourse
2. `timestamp` which contains the fetched version in the following format: `2006-01-02 15:04:05.999999999 -0700 MST`
3. `epoch` which contains the fetched version as a Unix epoch Timestamp (integer only)

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

### Cron trigger

```yaml
resources:
- name: nightly-build
  type: time
  source:
    cron: "0 0 * * *"  # Every day at midnight

jobs:
- name: run-nightly-build
  plan:
  - get: nightly-build
    trigger: true
  - task: build
    config: # ...
```

### Cron trigger using tags

```yaml
resources:
- name: weekly-cleanup
  type: time
  source:
    cron: "@weekly"  # Every Sunday at midnight

jobs:
- name: run-weekly-cleanup
  plan:
  - get: weekly-cleanup
    trigger: true
  - task: cleanup
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

### Trigger on/after specific datetime set

```yaml
resources:
- name: time-based-resource
  type: time
  source:
    start_after: 2023-12-01T09:00:00
    location: America/New_York

jobs:
- name: process-time-based-resource
  plan:
  - get: time-based-resource
    trigger: true
  - task: process-data
    config: # ...
```

### Trigger only on specific days

```yaml
resources:
- name: weekday-mornings
  type: time
  source:
    interval: 1h
    start: 9:00 AM
    stop: 12:00 PM
    days: [Monday, Tuesday, Wednesday, Thursday, Friday]
    location: Europe/London

jobs:
- name: something-every-hour-on-weekday-mornings
  plan:
  - get: weekday-mornings
    trigger: true
  - task: something
    config: # ...
```

### Cron trigger with modifiers

```yaml
resources:
- name: last-day-of-month
  type: time
  source:
    cron: "0 9 L * *"  # 9:00 AM on the last day of each month
    location: America/New_York

jobs:
- name: monthly-report
  plan:
  - get: last-day-of-month
    trigger: true
  - task: generate-report
    config: # ...
```
```yaml
resources:
- name: second-monday
  type: time
  source:
    cron: "0 7 * * 1#2"  # 7:00 AM on the second Monday of each month
    location: Europe/Berlin

jobs:
- name: bi-monthly-planning
  plan:
  - get: second-monday
    trigger: true
  - task: planning-meeting
    config: # ...
```

## Development

### Prerequisites

* golang is *required* - version 1.22.x is tested; earlier versions may also
  work.
* docker is *required* - version 25.x is tested; earlier versions may also
  work.
* go mod is used for dependency management of the golang packages.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following command:

```sh
docker build -t time-resource --target tests .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.