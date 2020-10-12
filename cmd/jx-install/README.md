# jx-install

Verifies the Jenkins X installation is operating correctly

Checks include:
- Git operator is running
- Boot jobs are able to run
- Boot jobs are successful

## Environment variables

Used to configure health checks

`BOOT_JOB_HEALTH_TIME_EXCEEDED` - used to decide how long to wait before reporting a boot job is stuck 