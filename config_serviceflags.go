package evergreen

import (
	"github.com/evergreen-ci/evergreen/db"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

// ServiceFlags holds the state of each of the runner/API processes
type ServiceFlags struct {
	TaskDispatchDisabled         bool `bson:"task_dispatch_disabled" json:"task_dispatch_disabled"`
	HostinitDisabled             bool `bson:"hostinit_disabled" json:"hostinit_disabled"`
	MonitorDisabled              bool `bson:"monitor_disabled" json:"monitor_disabled"`
	AlertsDisabled               bool `bson:"alerts_disabled" json:"alerts_disabled"`
	TaskrunnerDisabled           bool `bson:"taskrunner_disabled" json:"taskrunner_disabled"`
	RepotrackerDisabled          bool `bson:"repotracker_disabled" json:"repotracker_disabled"`
	SchedulerDisabled            bool `bson:"scheduler_disabled" json:"scheduler_disabled"`
	GithubPRTestingDisabled      bool `bson:"github_pr_testing_disabled" json:"github_pr_testing_disabled"`
	RepotrackerPushEventDisabled bool `bson:"repotracker_push_event_disabled" json:"repotracker_push_event_disabled"`
	CLIUpdatesDisabled           bool `bson:"cli_updates_disabled" json:"cli_updates_disabled"`
	BackgroundStatsDisabled      bool `bson:"background_stats_disabled" json:"background_stats_disabled"`
	TaskLoggingDisabled          bool `bson:"task_logging_disabled" json:"task_logging_disabled"`
	CacheStatsJobDisabled        bool `bson:"cache_stats_job_disabled" json:"cache_stats_job_disabled"`

	// Notification Flags
	EventProcessingDisabled      bool `bson:"event_processing_disabled" json:"event_processing_disabled"`
	JIRANotificationsDisabled    bool `bson:"jira_notifications_disabled" json:"jira_notifications_disabled"`
	SlackNotificationsDisabled   bool `bson:"slack_notifications_disabled" json:"slack_notifications_disabled"`
	EmailNotificationsDisabled   bool `bson:"email_notifications_disabled" json:"email_notifications_disabled"`
	WebhookNotificationsDisabled bool `bson:"webhook_notifications_disabled" json:"webhook_notifications_disabled"`
	GithubStatusAPIDisabled      bool `bson:"github_status_api_disabled" json:"github_status_api_disabled"`
}

func (c *ServiceFlags) SectionId() string { return "service_flags" }

func (c *ServiceFlags) Get() error {
	err := db.FindOneQ(ConfigCollection, db.Query(byId(c.SectionId())), c)
	if err != nil && err.Error() == errNotFound {
		*c = ServiceFlags{}
		return nil
	}
	return errors.Wrapf(err, "error retrieving section %s", c.SectionId())
}

func (c *ServiceFlags) Set() error {
	_, err := db.Upsert(ConfigCollection, byId(c.SectionId()), bson.M{
		"$set": bson.M{
			taskDispatchKey:                 c.TaskDispatchDisabled,
			hostinitKey:                     c.HostinitDisabled,
			monitorKey:                      c.MonitorDisabled,
			alertsKey:                       c.AlertsDisabled,
			taskrunnerKey:                   c.TaskrunnerDisabled,
			repotrackerKey:                  c.RepotrackerDisabled,
			schedulerKey:                    c.SchedulerDisabled,
			githubPRTestingDisabledKey:      c.GithubPRTestingDisabled,
			repotrackerPushEventDisabledKey: c.RepotrackerPushEventDisabled,
			cliUpdatesDisabledKey:           c.CLIUpdatesDisabled,
			backgroundStatsDisabledKey:      c.BackgroundStatsDisabled,
			eventProcessingDisabledKey:      c.EventProcessingDisabled,
			jiraNotificationsDisabledKey:    c.JIRANotificationsDisabled,
			slackNotificationsDisabledKey:   c.SlackNotificationsDisabled,
			emailNotificationsDisabledKey:   c.EmailNotificationsDisabled,
			webhookNotificationsDisabledKey: c.WebhookNotificationsDisabled,
			githubStatusAPIDisabledKey:      c.GithubStatusAPIDisabled,
			taskLoggingDisabledKey:          c.TaskLoggingDisabled,
			cacheStatsJobDisabledKey:        c.CacheStatsJobDisabled,
		},
	})
	return errors.Wrapf(err, "error updating section %s", c.SectionId())
}

func (c *ServiceFlags) ValidateAndDefault() error { return nil }
