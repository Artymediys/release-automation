package branch_ops

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
)

func MergeBranches(glc *gitlab.Client, projectID, sourceBranch, targetBranch string) error {
	mergeRequest, err := createMergeRequest(glc, projectID, sourceBranch, targetBranch)
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	err = cancelPipelines(glc, projectID)
	if err != nil {
		return err
	}

	err = acceptMergeRequest(glc, projectID, mergeRequest.IID)
	if err != nil {
		return err
	}

	return nil
}

func createMergeRequest(glc *gitlab.Client, projectID, sourceBranch, targetBranch string) (*gitlab.MergeRequest, error) {
	title := fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch)
	description := title
	rmSrcBranch := false

	opts := &gitlab.CreateMergeRequestOptions{
		SourceBranch:       &sourceBranch,
		TargetBranch:       &targetBranch,
		Title:              &title,
		Description:        &description,
		RemoveSourceBranch: &rmSrcBranch,
	}

	mergeRequest, _, err := glc.MergeRequests.CreateMergeRequest(projectID, opts)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Merge Request -> %w", err)
	}

	return mergeRequest, nil
}

func cancelPipelines(glc *gitlab.Client, projectID string) error {
	sourcePipeline := "merge_request_event"
	pipelines, _, err := glc.Pipelines.ListProjectPipelines(projectID, &gitlab.ListProjectPipelinesOptions{
		Source: &sourcePipeline,
		//Username: &gitlabUser,
	})
	if err != nil {
		return fmt.Errorf("не удалось получить список пайплайнов -> %w", err)
	}

	for _, pipeline := range pipelines {
		_, _, err = glc.Pipelines.CancelPipelineBuild(projectID, pipeline.ID)
		if err != nil {
			return fmt.Errorf("не удалось отменить пайплайн с ID %d -> %w", pipeline.ID, err)
		}
	}

	return nil
}

func acceptMergeRequest(glc *gitlab.Client, projectID string, mergeRequestID int) error {
	forceMerge := false
	opts := &gitlab.AcceptMergeRequestOptions{
		MergeWhenPipelineSucceeds: &forceMerge,
	}

	_, _, err := glc.MergeRequests.AcceptMergeRequest(projectID, mergeRequestID, opts)
	if err != nil {
		return fmt.Errorf("не удалось подтвердить Merge Request -> %w", err)
	}

	return nil
}
