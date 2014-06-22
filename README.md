# Time Resource

Implements a resource that reports new versions on a configured interval. The
interval can be arbitrarily long.

This can be used to ensure that a job executes a build at least once every N
minutes. The version tracked by this resource is simply the current time. It
does this by having `/in` report the current time as its resulting version, and
`/check` just compares the current time to the version passed in, and reports a
new version if `current time - previous time` is larger than `interval`.

See the tests under `check/` and `in/`.
