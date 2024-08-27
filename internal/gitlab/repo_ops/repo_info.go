package repo_ops

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

func GetGroupProjectMap(glc *gitlab.Client) (map[string][]string, error) {
	groupProjectMap := make(map[string][]string, 100)

	groups, err := getGroupsAndSubgroups(glc)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		projects, err := getProjectsInGroup(glc, group.ID)
		if err != nil {
			return nil, err
		}

		projectNames := make([]string, 0, 50)
		for _, project := range projects {
			projectNames = append(projectNames, project.Name)
		}

		groupProjectMap[group.Name] = projectNames
	}

	return groupProjectMap, nil
}

func GetProjectID(glc *gitlab.Client, projectName string) (string, error) {
	projects, _, err := glc.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Search: &projectName,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось получить список проектов -> %w", err)
	}

	for _, project := range projects {
		if project.Name == projectName {
			return strconv.Itoa(project.ID), nil
		}
	}

	return "", fmt.Errorf("не удалось найти указанный проект -> %s", projectName)
}

func GetProjectURL(glc *gitlab.Client, projectID string) (string, string, error) {
	project, _, err := glc.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
	if err != nil {
		return "", "", fmt.Errorf("не удалось получить информацию о проеке -> %w", err)
	}

	projectUrl, err := url.Parse(project.HTTPURLToRepo)
	if err != nil {
		return "", "", fmt.Errorf("не удалось распарсить URL проекта -> %w", err)
	}

	projectUrlBody := strings.TrimPrefix(project.HTTPURLToRepo, projectUrl.Scheme+"://")

	return projectUrl.Scheme, projectUrlBody, nil
}

func GetGroupNames(gls *gitlab.Client) ([]string, error) {
	groups, err := getGroupsAndSubgroups(gls)
	if err != nil {
		return nil, err
	}

	groupNames := make([]string, 0, 100)
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
	}

	return groupNames, nil
}

func getGroupsAndSubgroups(glc *gitlab.Client) ([]*gitlab.Group, error) {
	allGroupsMap := make(map[int]*gitlab.Group, 50)
	opt := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		groups, resp, err := glc.Groups.ListGroups(opt)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить список групп -> %w", err)
		}

		for _, group := range groups {
			if _, exists := allGroupsMap[group.ID]; !exists {
				allGroupsMap[group.ID] = group
				subgroups, err := getSubgroups(glc, group.ID, allGroupsMap)
				if err != nil {
					return nil, fmt.Errorf("не удалось получить подгруппы для группы %s -> %w", group.Name, err)
				}
				for _, subgroup := range subgroups {
					allGroupsMap[subgroup.ID] = subgroup
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	allGroups := make([]*gitlab.Group, 0, len(allGroupsMap))
	for _, group := range allGroupsMap {
		allGroups = append(allGroups, group)
	}

	return allGroups, nil
}

func getSubgroups(glc *gitlab.Client, groupID int, allGroupsMap map[int]*gitlab.Group) ([]*gitlab.Group, error) {
	allSubgroups := make([]*gitlab.Group, 0, 5)
	opt := &gitlab.ListSubGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		subgroups, resp, err := glc.Groups.ListSubGroups(groupID, opt)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить подгруппы -> %w", err)
		}

		for _, subgroup := range subgroups {
			if _, exists := allGroupsMap[subgroup.ID]; !exists {
				allSubgroups = append(allSubgroups, subgroup)
				allGroupsMap[subgroup.ID] = subgroup
				subsubgroups, err := getSubgroups(glc, subgroup.ID, allGroupsMap)
				if err != nil {
					return nil, fmt.Errorf("не удалось получить подгруппы для подгруппы %s -> %w", subgroup.Name, err)
				}
				allSubgroups = append(allSubgroups, subsubgroups...)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allSubgroups, nil
}

func getProjectsInGroup(glc *gitlab.Client, groupID int) ([]*gitlab.Project, error) {
	allProjects := make([]*gitlab.Project, 0, 5)
	opt := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	for {
		projects, resp, err := glc.Groups.ListGroupProjects(groupID, opt)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить список репозиториев в группе -> %w", err)
		}

		allProjects = append(allProjects, projects...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allProjects, nil
}

//func GetProjectIDMap(glc *gitlab.Client) (map[string]int, error) {
//	projectIDMap := make(map[string]int, 100)
//
//	groups, err := getGroupsAndSubgroups(glc)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, group := range groups {
//		projects, err := getProjectsInGroup(glc, group.ID)
//		if err != nil {
//			return nil, err
//		}
//
//		for _, project := range projects {
//			projectIDMap[project.Name] = project.ID
//		}
//	}
//
//	return projectIDMap, nil
//}
