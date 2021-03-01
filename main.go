package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/dougEfresh/gtoggl"
	gttimeentry "github.com/dougEfresh/gtoggl-api/gttimentry"
	log "github.com/sirupsen/logrus"
	"gopkg.in/andygrunwald/go-jira.v1"
	"os"
	"os/user"
	"strings"
)

type settings struct {
	jiraName     string
	jiraPassword string
	jiraUrl      string
	jiraQuery    string
	toggleToken  string
	dryRun       bool
}

func main() {
	verbose := flag.Bool("v", false, "verbose mode")
	dryRun := flag.Bool("dry-run", false, "dry run")
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	config, err := readSettings()
	if err != nil {
		panic(err)
	}
	config.dryRun = *dryRun

	tp := jira.BasicAuthTransport{Username: config.jiraName, Password: config.jiraPassword}
	jiraClient, err := jira.NewClient(tp.Client(), config.jiraUrl)
	if err != nil {
		panic(err)
	}
	issueService := jiraClient.Issue
	issues, _, err := issueService.Search(config.jiraQuery, nil)
	if err != nil {
		panic(err)
	}

	toggle, err := gtoggl.NewClient(config.toggleToken)
	if err != nil {
		panic(err)
	}
	worklogs, err := toggle.TimeentryClient.List()
	if err != nil {
		panic(err)
	}

	log.Infof("There are %d jira issues to check within last %d worklogs in toggle\n", len(issues), len(worklogs))

	for _, i := range issues {
		log.Debugf("Jira issue '%s' is in progress", i.Key)
		for _, w := range worklogs {
			if w.Duration <= 0 {
				continue
			}

			log.Debugf("-> Checking worklog '%s' for jira issue '%s'", w.Description, i.Key)
			if strings.Contains(w.Description, i.Key) {
				err := processIssue(config, issueService, i, w)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func processIssue(config settings, issueService *jira.IssueService, issue jira.Issue, worklog gttimeentry.TimeEntry) error {
	url := config.jiraUrl + "browse/" + issue.Key
	logs, _, err := issueService.GetWorklogs(issue.Key)
	if err != nil {
		return err
	}

	comment := fmt.Sprintf("toggl#%d", worklog.Id)
	for _, l := range logs.Worklogs {
		if l.TimeSpentSeconds == int(worklog.Duration) || l.Comment == comment {
			return nil
		}
	}
	log.Infof("There are %d minutes for an issue %s\n", int(worklog.Duration/60), url)
	if config.dryRun {
		return nil
	}

	_, _, err = issueService.AddWorklogRecord(issue.Key, &jira.WorklogRecord{
		TimeSpentSeconds: int(worklog.Duration), Comment: comment})
	return err
}

func readSettings() (result settings, err error) {
	usr, err := user.Current()
	if err != nil {
		return result, err
	}

	file, err := os.Open(usr.HomeDir + "/.toggl2jira")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	result.jiraUrl = strings.TrimSpace(scanner.Text())
	scanner.Scan()
	result.jiraName = strings.TrimSpace(scanner.Text())
	scanner.Scan()
	result.jiraPassword = strings.TrimSpace(scanner.Text())
	scanner.Scan()
	result.toggleToken = strings.TrimSpace(scanner.Text())
	scanner.Scan()
	result.jiraQuery = "assignee = currentUser() and (status = 'In Review' or status = 'In progress') order by updated desc"
	if scanner.Text() != "" {
		result.jiraQuery = strings.TrimSpace(scanner.Text())
	}

	return result, nil
}
