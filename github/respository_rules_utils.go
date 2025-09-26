package github

import (
	"reflect"
	"sort"

	"github.com/google/go-github/v74/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRulesetObject(d *schema.ResourceData, org string) *github.RepositoryRuleset {
	isOrgLevel := len(org) > 0

	var source, sourceType string
	if isOrgLevel {
		source = org
		sourceType = "Organization"
	} else {
		source = d.Get("repository").(string)
		sourceType = "Repository"
	}

	target := github.RulesetTarget(d.Get("target").(string))
	sourceTypeEnum := github.RulesetSourceType(sourceType)
	enforcement := github.RulesetEnforcement(d.Get("enforcement").(string))
	
	return &github.RepositoryRuleset{
		Name:         d.Get("name").(string),
		Target:       &target,
		Source:       source,
		SourceType:   &sourceTypeEnum,
		Enforcement:  enforcement,
		BypassActors: expandBypassActors(d.Get("bypass_actors").([]interface{})),
		Conditions:   expandConditions(d.Get("conditions").([]interface{}), isOrgLevel),
		Rules:        expandRules(d.Get("rules").([]interface{}), isOrgLevel),
	}
}

func expandBypassActors(input []interface{}) []*github.BypassActor {
	if len(input) == 0 {
		return nil
	}
	bypassActors := make([]*github.BypassActor, 0)

	for _, v := range input {
		inputMap := v.(map[string]interface{})
		actor := &github.BypassActor{}
		if v, ok := inputMap["actor_id"].(int); ok {
			actor.ActorID = github.Int64(int64(v))
		}

		if v, ok := inputMap["actor_type"].(string); ok {
			actorType := github.BypassActorType(v)
			actor.ActorType = &actorType
		}

		if v, ok := inputMap["bypass_mode"].(string); ok {
			bypassMode := github.BypassMode(v)
			actor.BypassMode = &bypassMode
		}
		bypassActors = append(bypassActors, actor)
	}
	return bypassActors
}

func flattenBypassActors(bypassActors []*github.BypassActor) []interface{} {
	if bypassActors == nil {
		return []interface{}{}
	}

	actorsSlice := make([]interface{}, 0)
	for _, v := range bypassActors {
		actorMap := make(map[string]interface{})

		actorMap["actor_id"] = v.GetActorID()
		actorMap["actor_type"] = v.GetActorType()
		actorMap["bypass_mode"] = v.GetBypassMode()

		actorsSlice = append(actorsSlice, actorMap)
	}

	return actorsSlice
}

func expandConditions(input []interface{}, org bool) *github.RepositoryRulesetConditions {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	rulesetConditions := &github.RepositoryRulesetConditions{}
	inputConditions := input[0].(map[string]interface{})

	// ref_name is available for both repo and org rulesets
	if v, ok := inputConditions["ref_name"].([]interface{}); ok && v != nil && len(v) != 0 {
		inputRefName := v[0].(map[string]interface{})
		include := make([]string, 0)
		exclude := make([]string, 0)

		for _, v := range inputRefName["include"].([]interface{}) {
			if v != nil {
				include = append(include, v.(string))
			}
		}

		for _, v := range inputRefName["exclude"].([]interface{}) {
			if v != nil {
				exclude = append(exclude, v.(string))
			}
		}

		rulesetConditions.RefName = &github.RepositoryRulesetRefConditionParameters{
			Include: include,
			Exclude: exclude,
		}
	}

	// org-only fields
	if org {
		// repository_name and repository_id
		if v, ok := inputConditions["repository_name"].([]interface{}); ok && v != nil && len(v) != 0 {
			inputRepositoryName := v[0].(map[string]interface{})
			include := make([]string, 0)
			exclude := make([]string, 0)

			for _, v := range inputRepositoryName["include"].([]interface{}) {
				if v != nil {
					include = append(include, v.(string))
				}
			}

			for _, v := range inputRepositoryName["exclude"].([]interface{}) {
				if v != nil {
					exclude = append(exclude, v.(string))
				}
			}

			protected := inputRepositoryName["protected"].(bool)

			rulesetConditions.RepositoryName = &github.RepositoryRulesetRepositoryNamesConditionParameters{
				Include:   include,
				Exclude:   exclude,
				Protected: &protected,
			}
		} else if v, ok := inputConditions["repository_id"].([]interface{}); ok && v != nil && len(v) != 0 {
			repositoryIDs := make([]int64, 0)

			for _, v := range v {
				if v != nil {
					repositoryIDs = append(repositoryIDs, int64(v.(int)))
				}
			}

			rulesetConditions.RepositoryID = &github.RepositoryRulesetRepositoryIDsConditionParameters{RepositoryIDs: repositoryIDs}
		}
	}

	return rulesetConditions
}

func flattenConditions(conditions *github.RepositoryRulesetConditions, org bool) []interface{} {
	if conditions == nil || conditions.RefName == nil {
		return []interface{}{}
	}

	conditionsMap := make(map[string]interface{})
	refNameSlice := make([]map[string]interface{}, 0)

	refNameSlice = append(refNameSlice, map[string]interface{}{
		"include": conditions.RefName.Include,
		"exclude": conditions.RefName.Exclude,
	})

	conditionsMap["ref_name"] = refNameSlice

	// org-only fields
	if org {
		repositoryNameSlice := make([]map[string]interface{}, 0)

		if conditions.RepositoryName != nil {
			var protected bool

			if conditions.RepositoryName.Protected != nil {
				protected = *conditions.RepositoryName.Protected
			}

			repositoryNameSlice = append(repositoryNameSlice, map[string]interface{}{
				"include":   conditions.RepositoryName.Include,
				"exclude":   conditions.RepositoryName.Exclude,
				"protected": protected,
			})
			conditionsMap["repository_name"] = repositoryNameSlice
		}

		if conditions.RepositoryID != nil {
			conditionsMap["repository_id"] = conditions.RepositoryID.RepositoryIDs
		}
	}

	return []interface{}{conditionsMap}
}

func expandRules(input []interface{}, org bool) *github.RepositoryRulesetRules {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	rulesMap := input[0].(map[string]interface{})
	rules := &github.RepositoryRulesetRules{}

	// Rules without parameters
	if v, ok := rulesMap["creation"].(bool); ok && v {
		rules.Creation = &github.EmptyRuleParameters{}
	}

	if v, ok := rulesMap["update"].(bool); ok && v {
		params := &github.UpdateRuleParameters{}
		if fetchAndMerge, ok := rulesMap["update_allows_fetch_and_merge"].(bool); ok {
			params.UpdateAllowsFetchAndMerge = fetchAndMerge
		}
		rules.Update = params
	}

	if v, ok := rulesMap["deletion"].(bool); ok && v {
		rules.Deletion = &github.EmptyRuleParameters{}
	}

	if v, ok := rulesMap["required_linear_history"].(bool); ok && v {
		rules.RequiredLinearHistory = &github.EmptyRuleParameters{}
	}

	if v, ok := rulesMap["required_signatures"].(bool); ok && v {
		rules.RequiredSignatures = &github.EmptyRuleParameters{}
	}

	if v, ok := rulesMap["non_fast_forward"].(bool); ok && v {
		rules.NonFastForward = &github.EmptyRuleParameters{}
	}

	// Required deployments rule (only for repository-level rulesets)
	if !org {
		if v, ok := rulesMap["required_deployments"].([]interface{}); ok && len(v) != 0 {
			requiredDeploymentsMap := make(map[string]interface{})
			if v[0] == nil {
				requiredDeploymentsMap["required_deployment_environments"] = make([]interface{}, 0)
			} else {
				requiredDeploymentsMap = v[0].(map[string]interface{})
			}
			envs := make([]string, 0)
			for _, v := range requiredDeploymentsMap["required_deployment_environments"].([]interface{}) {
				envs = append(envs, v.(string))
			}

			rules.RequiredDeployments = &github.RequiredDeploymentsRuleParameters{
				RequiredDeploymentEnvironments: envs,
			}
		}
	}

	// Pattern parameter rules
	if v, ok := rulesMap["commit_message_pattern"].([]interface{}); ok && len(v) != 0 {
		patternParametersMap := v[0].(map[string]interface{})
		rules.CommitMessagePattern = &github.PatternRuleParameters{
			Name:     github.String(patternParametersMap["name"].(string)),
			Negate:   github.Bool(patternParametersMap["negate"].(bool)),
			Operator: github.PatternRuleOperator(patternParametersMap["operator"].(string)),
			Pattern:  patternParametersMap["pattern"].(string),
		}
	}

	if v, ok := rulesMap["commit_author_email_pattern"].([]interface{}); ok && len(v) != 0 {
		patternParametersMap := v[0].(map[string]interface{})
		rules.CommitAuthorEmailPattern = &github.PatternRuleParameters{
			Name:     github.String(patternParametersMap["name"].(string)),
			Negate:   github.Bool(patternParametersMap["negate"].(bool)),
			Operator: github.PatternRuleOperator(patternParametersMap["operator"].(string)),
			Pattern:  patternParametersMap["pattern"].(string),
		}
	}

	if v, ok := rulesMap["committer_email_pattern"].([]interface{}); ok && len(v) != 0 {
		patternParametersMap := v[0].(map[string]interface{})
		rules.CommitterEmailPattern = &github.PatternRuleParameters{
			Name:     github.String(patternParametersMap["name"].(string)),
			Negate:   github.Bool(patternParametersMap["negate"].(bool)),
			Operator: github.PatternRuleOperator(patternParametersMap["operator"].(string)),
			Pattern:  patternParametersMap["pattern"].(string),
		}
	}

	if v, ok := rulesMap["branch_name_pattern"].([]interface{}); ok && len(v) != 0 {
		patternParametersMap := v[0].(map[string]interface{})
		rules.BranchNamePattern = &github.PatternRuleParameters{
			Name:     github.String(patternParametersMap["name"].(string)),
			Negate:   github.Bool(patternParametersMap["negate"].(bool)),
			Operator: github.PatternRuleOperator(patternParametersMap["operator"].(string)),
			Pattern:  patternParametersMap["pattern"].(string),
		}
	}

	if v, ok := rulesMap["tag_name_pattern"].([]interface{}); ok && len(v) != 0 {
		patternParametersMap := v[0].(map[string]interface{})
		rules.TagNamePattern = &github.PatternRuleParameters{
			Name:     github.String(patternParametersMap["name"].(string)),
			Negate:   github.Bool(patternParametersMap["negate"].(bool)),
			Operator: github.PatternRuleOperator(patternParametersMap["operator"].(string)),
			Pattern:  patternParametersMap["pattern"].(string),
		}
	}

	// Pull request rule
	if v, ok := rulesMap["pull_request"].([]interface{}); ok && len(v) != 0 {
		pullRequestMap := v[0].(map[string]interface{})
		rules.PullRequest = &github.PullRequestRuleParameters{
			DismissStaleReviewsOnPush:      pullRequestMap["dismiss_stale_reviews_on_push"].(bool),
			RequireCodeOwnerReview:         pullRequestMap["require_code_owner_review"].(bool),
			RequireLastPushApproval:        pullRequestMap["require_last_push_approval"].(bool),
			RequiredApprovingReviewCount:   pullRequestMap["required_approving_review_count"].(int),
			RequiredReviewThreadResolution: pullRequestMap["required_review_thread_resolution"].(bool),
		}
	}

	// Merge queue rule
	if v, ok := rulesMap["merge_queue"].([]interface{}); ok && len(v) != 0 {
		mergeQueueMap := v[0].(map[string]interface{})
		rules.MergeQueue = &github.MergeQueueRuleParameters{
			CheckResponseTimeoutMinutes:  mergeQueueMap["check_response_timeout_minutes"].(int),
			GroupingStrategy:             github.MergeGroupingStrategy(mergeQueueMap["grouping_strategy"].(string)),
			MaxEntriesToBuild:            mergeQueueMap["max_entries_to_build"].(int),
			MaxEntriesToMerge:            mergeQueueMap["max_entries_to_merge"].(int),
			MergeMethod:                  github.MergeQueueMergeMethod(mergeQueueMap["merge_method"].(string)),
			MinEntriesToMerge:            mergeQueueMap["min_entries_to_merge"].(int),
			MinEntriesToMergeWaitMinutes: mergeQueueMap["min_entries_to_merge_wait_minutes"].(int),
		}
	}

	// Required status checks rule
	if v, ok := rulesMap["required_status_checks"].([]interface{}); ok && len(v) != 0 {
		requiredStatusMap := v[0].(map[string]interface{})
		requiredStatusChecks := make([]*github.RuleStatusCheck, 0)

		if requiredStatusChecksInput, ok := requiredStatusMap["required_check"]; ok {
			requiredStatusChecksSet := requiredStatusChecksInput.(*schema.Set)
			for _, checkMap := range requiredStatusChecksSet.List() {
				check := checkMap.(map[string]interface{})
				params := &github.RuleStatusCheck{
					Context: check["context"].(string),
				}

				if integrationID := check["integration_id"].(int); integrationID != 0 {
					params.IntegrationID = github.Int64(int64(integrationID))
				}

				requiredStatusChecks = append(requiredStatusChecks, params)
			}
		}

		rules.RequiredStatusChecks = &github.RequiredStatusChecksRuleParameters{
			RequiredStatusChecks:             requiredStatusChecks,
			StrictRequiredStatusChecksPolicy: requiredStatusMap["strict_required_status_checks_policy"].(bool),
			DoNotEnforceOnCreate:             github.Bool(requiredStatusMap["do_not_enforce_on_create"].(bool)),
		}
	}

	// Required workflows to pass before merging rule
	if v, ok := rulesMap["required_workflows"].([]interface{}); ok && len(v) != 0 {
		requiredWorkflowsMap := v[0].(map[string]interface{})
		requiredWorkflows := make([]*github.RuleWorkflow, 0)

		if requiredWorkflowsInput, ok := requiredWorkflowsMap["required_workflow"]; ok {
			requiredWorkflowsSet := requiredWorkflowsInput.(*schema.Set)
			for _, workflowMap := range requiredWorkflowsSet.List() {
				workflow := workflowMap.(map[string]interface{})
				params := &github.RuleWorkflow{
					RepositoryID: github.Int64(int64(workflow["repository_id"].(int))),
					Path:         workflow["path"].(string),
					Ref:          github.String(workflow["ref"].(string)),
				}
				requiredWorkflows = append(requiredWorkflows, params)
			}
		}

		rules.Workflows = &github.WorkflowsRuleParameters{
			Workflows: requiredWorkflows,
		}
	}

	// Required code scanning to pass before merging rule
	if v, ok := rulesMap["required_code_scanning"].([]interface{}); ok && len(v) != 0 {
		requiredCodeScanningMap := v[0].(map[string]interface{})
		requiredCodeScanningTools := make([]*github.RuleCodeScanningTool, 0)

		if requiredCodeScanningInput, ok := requiredCodeScanningMap["required_code_scanning_tool"]; ok {
			requiredCodeScanningSet := requiredCodeScanningInput.(*schema.Set)
			for _, codeScanningMap := range requiredCodeScanningSet.List() {
				codeScanningTool := codeScanningMap.(map[string]interface{})
				params := &github.RuleCodeScanningTool{
					AlertsThreshold:         github.CodeScanningAlertsThreshold(codeScanningTool["alerts_threshold"].(string)),
					SecurityAlertsThreshold: github.CodeScanningSecurityAlertsThreshold(codeScanningTool["security_alerts_threshold"].(string)),
					Tool:                    codeScanningTool["tool"].(string),
				}
				requiredCodeScanningTools = append(requiredCodeScanningTools, params)
			}
		}

		rules.CodeScanning = &github.CodeScanningRuleParameters{
			CodeScanningTools: requiredCodeScanningTools,
		}
	}

	return rules
}

func flattenRules(rules *github.RepositoryRulesetRules, org bool) []interface{} {
	if rules == nil {
		return []interface{}{}
	}

	rulesMap := make(map[string]interface{})

	// Simple rules without parameters
	if rules.Creation != nil {
		rulesMap["creation"] = true
	}

	if rules.Update != nil {
		rulesMap["update"] = true
		rulesMap["update_allows_fetch_and_merge"] = rules.Update.UpdateAllowsFetchAndMerge
	}

	if rules.Deletion != nil {
		rulesMap["deletion"] = true
	}

	if rules.RequiredLinearHistory != nil {
		rulesMap["required_linear_history"] = true
	}

	if rules.RequiredSignatures != nil {
		rulesMap["required_signatures"] = true
	}

	if rules.NonFastForward != nil {
		rulesMap["non_fast_forward"] = true
	}

	// Required deployments rule (only for repository-level rulesets)
	if !org && rules.RequiredDeployments != nil {
		rule := make(map[string]interface{})
		rule["required_deployment_environments"] = rules.RequiredDeployments.RequiredDeploymentEnvironments
		rulesMap["required_deployments"] = []map[string]interface{}{rule}
	}

	// Pattern parameter rules
	if rules.CommitMessagePattern != nil {
		rule := make(map[string]interface{})
		rule["name"] = rules.CommitMessagePattern.GetName()
		rule["negate"] = rules.CommitMessagePattern.GetNegate()
		rule["operator"] = string(rules.CommitMessagePattern.Operator)
		rule["pattern"] = rules.CommitMessagePattern.Pattern
		rulesMap["commit_message_pattern"] = []map[string]interface{}{rule}
	}

	if rules.CommitAuthorEmailPattern != nil {
		rule := make(map[string]interface{})
		rule["name"] = rules.CommitAuthorEmailPattern.GetName()
		rule["negate"] = rules.CommitAuthorEmailPattern.GetNegate()
		rule["operator"] = string(rules.CommitAuthorEmailPattern.Operator)
		rule["pattern"] = rules.CommitAuthorEmailPattern.Pattern
		rulesMap["commit_author_email_pattern"] = []map[string]interface{}{rule}
	}

	if rules.CommitterEmailPattern != nil {
		rule := make(map[string]interface{})
		rule["name"] = rules.CommitterEmailPattern.GetName()
		rule["negate"] = rules.CommitterEmailPattern.GetNegate()
		rule["operator"] = string(rules.CommitterEmailPattern.Operator)
		rule["pattern"] = rules.CommitterEmailPattern.Pattern
		rulesMap["committer_email_pattern"] = []map[string]interface{}{rule}
	}

	if rules.BranchNamePattern != nil {
		rule := make(map[string]interface{})
		rule["name"] = rules.BranchNamePattern.GetName()
		rule["negate"] = rules.BranchNamePattern.GetNegate()
		rule["operator"] = string(rules.BranchNamePattern.Operator)
		rule["pattern"] = rules.BranchNamePattern.Pattern
		rulesMap["branch_name_pattern"] = []map[string]interface{}{rule}
	}

	if rules.TagNamePattern != nil {
		rule := make(map[string]interface{})
		rule["name"] = rules.TagNamePattern.GetName()
		rule["negate"] = rules.TagNamePattern.GetNegate()
		rule["operator"] = string(rules.TagNamePattern.Operator)
		rule["pattern"] = rules.TagNamePattern.Pattern
		rulesMap["tag_name_pattern"] = []map[string]interface{}{rule}
	}

	// Pull request rule
	if rules.PullRequest != nil {
		rule := make(map[string]interface{})
		rule["dismiss_stale_reviews_on_push"] = rules.PullRequest.DismissStaleReviewsOnPush
		rule["require_code_owner_review"] = rules.PullRequest.RequireCodeOwnerReview
		rule["require_last_push_approval"] = rules.PullRequest.RequireLastPushApproval
		rule["required_approving_review_count"] = rules.PullRequest.RequiredApprovingReviewCount
		rule["required_review_thread_resolution"] = rules.PullRequest.RequiredReviewThreadResolution
		rulesMap["pull_request"] = []map[string]interface{}{rule}
	}

	// Required status checks rule
	if rules.RequiredStatusChecks != nil {
		requiredStatusChecksSlice := make([]map[string]interface{}, 0)
		for _, check := range rules.RequiredStatusChecks.RequiredStatusChecks {
			integrationID := int64(0)
			if check.IntegrationID != nil {
				integrationID = *check.IntegrationID
			}
			requiredStatusChecksSlice = append(requiredStatusChecksSlice, map[string]interface{}{
				"context":        check.Context,
				"integration_id": integrationID,
			})
		}

		rule := make(map[string]interface{})
		rule["required_check"] = requiredStatusChecksSlice
		rule["strict_required_status_checks_policy"] = rules.RequiredStatusChecks.StrictRequiredStatusChecksPolicy
		rule["do_not_enforce_on_create"] = rules.RequiredStatusChecks.GetDoNotEnforceOnCreate()
		rulesMap["required_status_checks"] = []map[string]interface{}{rule}
	}

	// Merge queue rule
	if rules.MergeQueue != nil {
		rule := make(map[string]interface{})
		rule["check_response_timeout_minutes"] = rules.MergeQueue.CheckResponseTimeoutMinutes
		rule["grouping_strategy"] = string(rules.MergeQueue.GroupingStrategy)
		rule["max_entries_to_build"] = rules.MergeQueue.MaxEntriesToBuild
		rule["max_entries_to_merge"] = rules.MergeQueue.MaxEntriesToMerge
		rule["merge_method"] = string(rules.MergeQueue.MergeMethod)
		rule["min_entries_to_merge"] = rules.MergeQueue.MinEntriesToMerge
		rule["min_entries_to_merge_wait_minutes"] = rules.MergeQueue.MinEntriesToMergeWaitMinutes
		rulesMap["merge_queue"] = []map[string]interface{}{rule}
	}

	// Required workflows rule
	if rules.Workflows != nil {
		requiredWorkflowsSlice := make([]map[string]interface{}, 0)
		for _, workflow := range rules.Workflows.Workflows {
			workflowMap := map[string]interface{}{
				"path":          workflow.Path,
				"repository_id": workflow.GetRepositoryID(),
				"ref":           workflow.GetRef(),
			}
			requiredWorkflowsSlice = append(requiredWorkflowsSlice, workflowMap)
		}

		rule := make(map[string]interface{})
		rule["required_workflow"] = requiredWorkflowsSlice
		rulesMap["required_workflows"] = []map[string]interface{}{rule}
	}

	// Required code scanning rule
	if rules.CodeScanning != nil {
		requiredCodeScanningSlice := make([]map[string]interface{}, 0)
		for _, codeScanningTool := range rules.CodeScanning.CodeScanningTools {
			codeScanningMap := map[string]interface{}{
				"alerts_threshold":          string(codeScanningTool.AlertsThreshold),
				"security_alerts_threshold": string(codeScanningTool.SecurityAlertsThreshold),
				"tool":                      codeScanningTool.Tool,
			}
			requiredCodeScanningSlice = append(requiredCodeScanningSlice, codeScanningMap)
		}

		rule := make(map[string]interface{})
		rule["required_code_scanning_tool"] = requiredCodeScanningSlice
		rulesMap["required_code_scanning"] = []map[string]interface{}{rule}
	}

	return []interface{}{rulesMap}
}

func bypassActorsDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// If the length has changed, no need to suppress
	if k == "bypass_actors.#" {
		return old == new
	}

	// Get change to bypass actors
	o, n := d.GetChange("bypass_actors")
	oldBypassActors := o.([]interface{})
	newBypassActors := n.([]interface{})

	sort.SliceStable(oldBypassActors, func(i, j int) bool {
		return oldBypassActors[i].(map[string]interface{})["actor_id"].(int) > oldBypassActors[j].(map[string]interface{})["actor_id"].(int)
	})
	sort.SliceStable(newBypassActors, func(i, j int) bool {
		return newBypassActors[i].(map[string]interface{})["actor_id"].(int) > newBypassActors[j].(map[string]interface{})["actor_id"].(int)
	})

	return reflect.DeepEqual(oldBypassActors, newBypassActors)
}
