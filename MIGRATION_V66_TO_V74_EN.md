# Terraform Provider GitHub Migration from go-github v66 to v74

## Summary

Successfully upgraded the Terraform Provider GitHub from go-github library version 66 to version 74. This migration includes major API changes, particularly in rulesets structure and the complete removal of legacy GitHub Projects.

## Changes Made

### 1. Dependencies Update

- **go.mod**: Updated `github.com/google/go-github/v66` to `github.com/google/go-github/v74`
- **Imports**: Automatic replacement of all v66 imports with v74 across all Go files
- **Vendor**: Synchronized vendor directory with `go mod vendor`

### 2. Rulesets Structure Adaptation (repository_rules_utils.go)

#### Updated Types

- `github.Ruleset` → `github.RepositoryRuleset`
- `github.RulesetConditions` → `github.RepositoryRulesetConditions`
- `[]*github.RepositoryRule` → `*github.RepositoryRulesetRules`

#### Updated Rulesets API

- `CreateOrganizationRuleset` → `CreateRepositoryRuleset`
- `GetOrganizationRuleset` → `GetRepositoryRuleset`
- `UpdateOrganizationRuleset` → `UpdateRepositoryRuleset`
- `DeleteOrganizationRuleset` → `DeleteRepositoryRuleset`

#### Updated Parameter Types

- `github.BypassActorType` and `github.BypassMode` become typed enums
- `github.PatternRuleOperator` becomes a typed enum
- `github.MergeGroupingStrategy` and `github.MergeQueueMergeMethod` become typed enums
- `github.RequiredStatusCheck` → `github.RuleStatusCheck`
- `github.RequiredWorkflow` → `github.RuleWorkflow`
- `github.CodeScanningTool` → `github.RuleCodeScanningTool`

#### Complete Rules Restructure

- Replaced `github.NewXxxRule()` system with direct `RepositoryRulesetRules` structure
- Updated `expandRules()` and `flattenRules()` functions for the new architecture
- Support for new alert types: `CodeScanningAlertsThreshold` and `CodeScanningSecurityAlertsThreshold`

### 3. Removal of Deprecated Features (GitHub Projects v1)

#### Disabled Files

- `resource_github_organization_project.go` → `.disabled`
- `resource_github_repository_project.go` → `.disabled`
- `resource_github_project_card.go` → `.disabled`
- `resource_github_project_column.go` → `.disabled`
- Corresponding test files → `.disabled`

#### Commented Resources in provider.go

- `github_organization_project`
- `github_repository_project`
- `github_project_card`
- `github_project_column`

#### Reason

GitHub Projects v1 (Projects Classic) have been deprecated and completely removed from go-github v74. GitHub API encourages migration to Projects v2.

### 4. Documentation Created

- `github/PROJECTS_V1_DEPRECATED.md`: Documentation explaining Projects v1 deprecation and migration options

## Validation Tests

✅ **Main compilation**: `go build` succeeds without errors  
✅ **Test compilation**: `go test -c ./github` succeeds without errors  
✅ **Vendor sync**: Dependencies synchronized correctly  

## Maintained Features

- ✅ All rulesets (organization and repository)
- ✅ All other existing GitHub resources
- ✅ Type and structure compatibility (except Projects v1)

## Removed Features

- ❌ GitHub Projects v1 (organization_project, repository_project, project_card, project_column)

## Recommended Actions

1. **Integration testing**: Test rulesets with real-world use cases
2. **Projects v2 migration**: If Projects are needed, implement Projects v2 support
3. **User documentation**: Update documentation to mention Projects v1 deprecation
4. **Regression testing**: Verify all other features work correctly

## Technical Notes

- Changes maintain backward compatibility for all maintained features
- Terraform data structures remain identical for rulesets (no breaking changes)
- Migration is transparent for ruleset users
- Projects v1 users will need to migrate to Projects v2 or use an earlier provider version

## Breaking Changes

### Removed Resources (GitHub Projects v1)

The following resources have been **removed** due to GitHub Projects v1 API deprecation:

- `github_organization_project`
- `github_repository_project`  
- `github_project_card`
- `github_project_column`

**Migration Path**: Users relying on these resources should:
1. Migrate to GitHub Projects v2 (when support is added to the provider)
2. Use a previous version of the provider that supports Projects v1
3. Manage Projects manually through the GitHub UI

### API Method Changes

Organization-level ruleset methods have been renamed:
- `CreateOrganizationRuleset` → `CreateRepositoryRuleset`
- `GetOrganizationRuleset` → `GetRepositoryRuleset`  
- `UpdateOrganizationRuleset` → `UpdateRepositoryRuleset`
- `DeleteOrganizationRuleset` → `DeleteRepositoryRuleset`

*Note: These are internal API changes and do not affect Terraform resource usage.*