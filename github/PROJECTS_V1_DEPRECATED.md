# GitHub Projects v1 Deprecated

## Context

During the upgrade from go-github v66 to v74, we discovered that the legacy GitHub Projects v1 API has been completely removed from the go-github library.

## Files Affected

- `resource_github_organization_project.go` → renamed to `resource_github_organization_project.go.disabled`

## Reason for Removal

GitHub has deprecated Projects v1 in favor of Projects v2 (also known as Projects Beta). The go-github library v74 no longer includes support for the legacy Projects API, which used:

- `client.Projects` service (removed)
- `github.ProjectOptions` type (removed)  
- `client.Organizations.CreateProject` method (removed)

## Next Steps

1. **Option 1: Remove completely** - If the organization projects feature is no longer needed
2. **Option 2: Migrate to Projects v2** - Implement support for the new Projects v2 API using:
   - ProjectsV2 service
   - New ProjectV2 types
   - Updated API endpoints

## Migration Guide

If migrating to Projects v2, you'll need to:

1. Use the new ProjectsV2 types and services
2. Update the Terraform schema to match Projects v2 capabilities
3. Migrate existing state to the new resource format
4. Update documentation and examples

## References

- [GitHub Projects v2 API Documentation](https://docs.github.com/en/rest/projects)
- [GitHub Projects v1 to v2 Migration Guide](https://docs.github.com/en/issues/planning-and-tracking-with-projects/creating-projects/migrating-from-projects-classic)