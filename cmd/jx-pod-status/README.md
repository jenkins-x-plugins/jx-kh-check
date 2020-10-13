# jx-pod-status

This is pretty much a copy from the official Kuberhealthy check for pod statuses see https://github.com/Comcast/kuberhealthy/tree/master/cmd/pod-status-check

The main difference is it supports being run with a cluster role and so will check pods in all namespaces.

The changes could be helm templated and submitted to the official chart pending the outcome of this discussion https://github.com/Comcast/kuberhealthy/issues/680