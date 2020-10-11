This is a collection of custom health checks used by Kuberhealthy

Health checks use the Kuberhealthy API to report errors.

A good practice for writing checks are:
 - keep them small so they use the smallest resources possible
 - keep them specific so that RBAC for service accounts can have reduced permissions
 - ensure helm chart can be enabled / disabled via helm values   