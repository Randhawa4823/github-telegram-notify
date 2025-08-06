package utils

import (
        "encoding/json"
        "fmt"
        "html"
        "strings"

        "github.com/google/go-github/v71/github"
)

// Helper functions
func formatText(input string, maxLen int) string {
        input = html.EscapeString(input)
        if maxLen > 0 && len(input) > maxLen {
                return input[:maxLen] + "..."
        }
        return input
}

func safeGetURL(url string) string {
        if url == "" {
                return "#"
        }
        return url
}

func truncateText(text string, maxLen int) string {
        if len(text) > maxLen {
                return text[:maxLen] + "..."
        }
        return text
}

func branchOrDefault(branch *string) string {
        if branch != nil {
                return *branch
        }
        return "default branch"
}

// Event handlers
func HandleIssuesEvent(event *github.IssuesEvent) string {
        repo := event.GetRepo().GetFullName()
        action := event.GetAction()
        sender := event.GetSender().GetLogin()
        issue := event.GetIssue()
        title := issue.GetTitle()
        url := issue.GetHTMLURL()

        msg := fmt.Sprintf(
                "<b>📌 %s issue</b>\n"+
                        "<b>Repo:</b> <a href='%s'>%s</a>\n"+
                        "<b>By:</b> %s\n",
                strings.Title(action),
                safeGetURL(url), repo,
                sender,
        )

        switch action {
        case "opened", "edited":
                msg += fmt.Sprintf("<b>Title:</b> %s\n", formatText(title, 0))
                if body := issue.GetBody(); body != "" {
                        msg += fmt.Sprintf("<b>Description:</b>\n%s\n", formatText(body, 300))
                }
        case "closed":
                if closer := issue.GetClosedBy(); closer != nil {
                        msg += fmt.Sprintf("<b>Closed by:</b> %s\n", closer.GetLogin())
                }
        case "reopened":
                msg += "<i>Issue reopened</i>\n"
        case "assigned":
                var assignees []string
                for _, a := range issue.Assignees {
                        assignees = append(assignees, a.GetLogin())
                }
                msg += fmt.Sprintf("<b>Assigned to:</b> %s\n", strings.Join(assignees, ", "))
        case "labeled":
                var labels []string
                for _, l := range issue.Labels {
                        labels = append(labels, l.GetName())
                }
                msg += fmt.Sprintf("<b>Labels:</b> %s\n", strings.Join(labels, ", "))
        case "milestoned":
                if m := issue.GetMilestone(); m != nil {
                        msg += fmt.Sprintf("<b>Milestone:</b> %s\n", m.GetTitle())
                }
        }

        msg += fmt.Sprintf("<a href='%s'>View Issue</a>", safeGetURL(url))
        return msg
}

func HandlePullRequestEvent(event *github.PullRequestEvent) string {
        repo := event.GetRepo().GetFullName()
        action := event.GetAction()
        sender := event.GetSender().GetLogin()
        pr := event.GetPullRequest()
        title := pr.GetTitle()
        url := pr.GetHTMLURL()
        state := pr.GetState()

        msg := fmt.Sprintf(
                "<b>🚀 PR %s</b>: <a href='%s'>%s</a>\n"+
                        "<b>Repo:</b> <a href='https://github.com/%s'>%s</a>\n"+
                        "<b>By:</b> %s | <b>State:</b> %s\n",
                strings.Title(action),
                safeGetURL(url), formatText(title, 100),
                repo, repo,
                sender, state,
        )

        switch action {
        case "opened":
                msg += fmt.Sprintf("<b>Description:</b>\n%s\n", formatText(pr.GetBody(), 300))
        case "closed":
                if pr.GetMerged() {
                        msg += "✅ Merged\n"
                } else {
                        msg += "❌ Closed without merging\n"
                }
        case "reopened":
                msg += "🔄 Reopened\n"
        case "edited":
                msg += fmt.Sprintf("✏️ Edited\n<b>Description:</b>\n%s\n", formatText(pr.GetBody(), 300))
        case "assigned":
                var assignees []string
                for _, a := range pr.Assignees {
                        assignees = append(assignees, a.GetLogin())
                }
                msg += fmt.Sprintf("<b>Assigned:</b> %s\n", strings.Join(assignees, ", "))
        case "review_requested":
                var reviewers []string
                for _, r := range pr.RequestedReviewers {
                        reviewers = append(reviewers, r.GetLogin())
                }
                msg += fmt.Sprintf("<b>Reviewers:</b> %s\n", strings.Join(reviewers, ", "))
        case "labeled":
                var labels []string
                for _, l := range pr.Labels {
                        labels = append(labels, l.GetName())
                }
                msg += fmt.Sprintf("<b>Labels:</b> %s\n", strings.Join(labels, ", "))
        case "synchronize":
                msg += "🔄 New commits pushed\n"
        }

        msg += fmt.Sprintf("<a href='%s'>View PR</a>", safeGetURL(url))
        return msg
}

func HandleStarredEvent(event *github.StarredRepository) string {
        repo := event.Repository.GetFullName()
        repoURL := event.Repository.GetHTMLURL()
        sender := event.Repository.Owner.GetLogin()
        stars := event.Repository.GetStargazersCount()

        return fmt.Sprintf(
                "🌟 <b>New star on</b> <a href='%s'>%s</a>\n"+
                        "👤 Starred by: %s\n"+
                        "✨ Total stars: %d",
                safeGetURL(repoURL),
                repo,
                sender,
                stars,
        )
}

func HandlePushEvent(event *github.PushEvent) string {
        repo := event.Repo.GetName()
        repoURL := event.Repo.GetHTMLURL()
        branch := strings.TrimPrefix(event.GetRef(), "refs/heads/")
        compareURL := event.GetCompare()
        commitCount := len(event.Commits)

        if commitCount == 0 {
                return ""
        }

        msg := fmt.Sprintf(
                "🔨 <b>%d</b> <a href='%s'>new commit(s)</a> to <code>%s:%s</code>:\n\n",
                commitCount, safeGetURL(compareURL), repo, branch,
        )

        if event.GetCreated() {
                msg += "🌱 <i>New branch created</i>\n"
        } else if event.GetDeleted() {
                msg += "🗑️ <i>Branch deleted</i>\n"
        } else if event.GetForced() {
                msg += "⚠️ <i>Force pushed</i>\n"
        }

        for i, commit := range event.Commits {
                if i >= 5 { // Limit to 5 commits
                        msg += fmt.Sprintf("• And %d more commits...\n", commitCount-5)
                        break
                }
                shortSHA := commit.GetID()
                if len(shortSHA) > 7 {
                        shortSHA = shortSHA[:7]
                }
                msg += fmt.Sprintf(
                        "• <a href='%s/commit/%s'>%s</a>: %s by %s\n",
                        safeGetURL(repoURL),
                        commit.GetID(),
                        shortSHA,
                        formatText(commit.GetMessage(), 80),
                        formatText(commit.Author.GetName(), 0),
                )
        }

        if len(msg) > 4000 {
                return fmt.Sprintf(
                        "🔨 <b>%d</b> <a href='%s'>new commit(s)</a> to <code>%s:%s</code>:\n\n"+
                                "⚠️ <i>Too many commits to display, check the repository for details.</i>\n",
                        commitCount, safeGetURL(compareURL), repo, branch,
                )
        }

        return msg
}

func HandleCreateEvent(event *github.CreateEvent) string {
        repo := event.Repo.GetFullName()
        repoURL := event.Repo.GetHTMLURL()
        sender := event.Sender.GetLogin()
        refType := event.GetRefType()
        ref := event.GetRef()

        msg := fmt.Sprintf(
                "✨ <b>New %s created</b>\n"+
                        "<b>Name:</b> <code>%s</code>\n"+
                        "<b>Repo:</b> <a href='%s'>%s</a>\n"+
                        "<b>By:</b> %s\n",
                refType,
                formatText(ref, 0),
                safeGetURL(repoURL),
                repo,
                sender,
        )

        if desc := event.GetDescription(); desc != "" {
                msg += fmt.Sprintf("<b>Description:</b> %s\n", formatText(desc, 0))
        }

        if refType == "repository" && event.GetMasterBranch() != "" {
                msg += fmt.Sprintf("<b>Default branch:</b> %s\n", event.GetMasterBranch())
        }

        return msg
}

func HandleDeleteEvent(event *github.DeleteEvent) string {
        repo := event.Repo.GetFullName()
        repoURL := event.Repo.GetHTMLURL()
        sender := event.Sender.GetLogin()
        refType := event.GetRefType()
        ref := event.GetRef()

        emoji := "❌"
        switch refType {
        case "branch":
                emoji = "🌿"
        case "tag":
                emoji = "🏷️"
        }

        return fmt.Sprintf(
                "%s <b>Deleted %s:</b> <code>%s</code>\n"+
                        "<b>Repo:</b> <a href='%s'>%s</a>\n"+
                        "<b>By:</b> %s",
                emoji,
                refType,
                formatText(ref, 0),
                safeGetURL(repoURL),
                repo,
                sender,
        )
}

func HandleForkEvent(event *github.ForkEvent) string {
        originalRepo := event.Repo.GetFullName()
        forkedRepo := event.Forkee.GetFullName()
        sender := event.Sender.GetLogin()
        return fmt.Sprintf(
                "🍴 <a href='https://github.com/%s'>%s</a> forked by %s\n"+
                        "✨ %d stars | 🍴 %d forks",
                forkedRepo,
                originalRepo,
                sender,
                event.Repo.GetStargazersCount(),
                event.Repo.GetForksCount(),
        )
}

func HandleCommitCommentEvent(event *github.CommitCommentEvent) string {
        comment := event.Comment.GetBody()
        commitSHA := event.Comment.GetCommitID()
        repo := event.Repo.GetFullName()
        sender := event.Sender.GetLogin()
        action := event.GetAction()
        commitURL := fmt.Sprintf("https://github.com/%s/commit/%s", repo, commitSHA)

        actionEmoji := map[string]string{
                "created": "💬",
                "edited":  "✏️",
                "deleted": "🗑️",
        }[action]

        if actionEmoji == "" {
                actionEmoji = "⚠️"
        }

        msg := fmt.Sprintf(
                "%s <b>%s</b> %s comment on commit\n"+
                        "<b>Repo:</b> <a href='https://github.com/%s'>%s</a>\n"+
                        "<b>Commit:</b> <a href='%s'>%s</a>\n",
                actionEmoji,
                sender,
                action,
                repo,
                repo,
                safeGetURL(commitURL),
                commitSHA[:7],
        )

        if action == "created" || action == "edited" {
                msg += fmt.Sprintf("<b>Comment:</b> %s", formatText(comment, 300))
        }

        return msg
}

func HandlePublicEvent(event *github.PublicEvent) string {
        return fmt.Sprintf(
                "🔓 <b>Repository made public</b>\n"+
                        "<b>Name:</b> <a href='%s'>%s</a>\n"+
                        "<b>By:</b> %s",
                safeGetURL(event.Repo.GetHTMLURL()),
                event.Repo.GetFullName(),
                event.Sender.GetLogin(),
        )
}

func HandleIssueCommentEvent(event *github.IssueCommentEvent) string {
        action := event.GetAction()
        issue := event.Issue
        comment := event.Comment
        repo := event.Repo.GetFullName()
        sender := event.Sender.GetLogin()

        actionEmoji := map[string]string{
                "created": "💬",
                "edited":  "✏️",
                "deleted": "🗑️",
        }[action]
        if actionEmoji == "" {
                actionEmoji = "⚠️"
        }

        msg := fmt.Sprintf(
                "%s <b>%s</b> %s comment on <a href='%s'>%s#%d</a>\n"+
                        "<b>Title:</b> %s\n",
                actionEmoji,
                sender,
                action,
                safeGetURL(issue.GetHTMLURL()),
                repo,
                issue.GetNumber(),
                formatText(issue.GetTitle(), 100),
        )

        if action == "created" || action == "edited" {
                msg += fmt.Sprintf("<b>Comment:</b> %s", formatText(comment.GetBody(), 300))
        }

        return msg
}

func HandleMemberEvent(event *github.MemberEvent) string {
        action := event.GetAction()
        member := event.Member.GetLogin()
        repo := event.Repo.GetFullName()
        org := event.Org.GetLogin()
        sender := event.Sender.GetLogin()

        actionInfo := map[string]struct {
                emoji string
                verb  string
        }{
                "added":   {"➕", "added to"},
                "removed": {"➖", "removed from"},
                "edited":  {"✏️", "updated in"},
        }[action]

        if actionInfo.emoji == "" {
                actionInfo = struct {
                        emoji string
                        verb  string
                }{"⚠️", "performed action on"}
        }

        msg := fmt.Sprintf(
                "%s <b>%s</b> %s %s/%s\n"+
                        "<b>By:</b> %s",
                actionInfo.emoji,
                member,
                actionInfo.verb,
                org,
                repo,
                sender,
        )

        if action == "edited" && event.Changes != nil {
                changes, _ := json.Marshal(event.Changes)
                msg += fmt.Sprintf("\n<b>Changes:</b> %s", formatText(string(changes), 100))
        }

        return msg
}

func HandleRepositoryEvent(event *github.RepositoryEvent) string {
        action := event.GetAction()
        repo := event.Repo.GetFullName()
        url := event.Repo.GetHTMLURL()
        sender := event.Sender.GetLogin()

        actionDetails := map[string]struct {
                emoji string
                desc  string
        }{
                "created":    {"🎉", "created"},
                "renamed":    {"🔄", fmt.Sprintf("renamed to %s", event.Repo.GetName())},
                "archived":   {"🔒", "archived"},
                "unarchived": {"🔓", "unarchived"},
        }[action]

        if actionDetails.emoji == "" {
                actionDetails = struct {
                        emoji string
                        desc  string
                }{"⚠️", fmt.Sprintf("performed %s action", action)}
        }

        return fmt.Sprintf(
                "%s <a href='%s'>%s</a> %s\n"+
                        "👤 By: %s",
                actionDetails.emoji,
                safeGetURL(url),
                repo,
                actionDetails.desc,
                sender,
        )
}

func HandleReleaseEvent(event *github.ReleaseEvent) string {
        action := event.GetAction()
        release := event.GetRelease()
        repo := event.GetRepo().GetFullName()
        sender := event.GetSender().GetLogin()

        actionDetails := map[string]struct {
                emoji string
                verb  string
        }{
                "created":   {"🎉", "New release"},
                "published": {"🚀", "Release published"},
                "deleted":   {"🗑️", "Release deleted"},
                "edited":    {"✏️", "Release edited"},
        }[action]

        if actionDetails.emoji == "" {
                actionDetails = struct {
                        emoji string
                        verb  string
                }{"⚠️", fmt.Sprintf("Unknown action (%s)", action)}
        }

        msg := fmt.Sprintf(
                "%s <b>%s</b> in <a href='%s'>%s</a>\n"+
                        "<b>Tag:</b> %s\n"+
                        "<b>By:</b> %s",
                actionDetails.emoji,
                actionDetails.verb,
                safeGetURL(release.GetHTMLURL()),
                repo,
                release.GetTagName(),
                sender,
        )

        if (action == "created" || action == "edited") && release.GetBody() != "" {
                msg += fmt.Sprintf("\n<b>Notes:</b> %s", formatText(release.GetBody(), 300))
        }

        return msg
}

func HandleWatchEvent(event *github.WatchEvent) string {
        action := event.GetAction()
        if action != "started" {
                return fmt.Sprintf(
                        "⚠️ Unexpected watch action: %s on %s by %s",
                        action,
                        event.GetRepo().GetFullName(),
                        event.GetSender().GetLogin(),
                )
        }
        return fmt.Sprintf(
                "⭐ %s starred %s",
                event.GetSender().GetLogin(),
                event.GetRepo().GetFullName(),
        )
}

func HandleStatusEvent(event *github.StatusEvent) string {
        state := event.GetState()
        stateEmoji := map[string]string{
                "success": "✅",
                "error":   "❌",
                "pending": "⏳",
        }[state]
        if stateEmoji == "" {
                stateEmoji = "⚠️"
        }

        return fmt.Sprintf(
                "%s <b>%s</b> for commit <a href='%s'>%s</a>\n"+
                        "<b>Repo:</b> <a href='%s'>%s</a>\n"+
                        "<b>Status:</b> %s\n"+
                        "<b>By:</b> %s",
                stateEmoji,
                strings.Title(state),
                safeGetURL(event.GetCommit().GetHTMLURL()),
                event.GetCommit().GetSHA()[:7],
                safeGetURL(event.GetRepo().GetHTMLURL()),
                event.GetRepo().GetFullName(),
                event.GetDescription(),
                event.GetSender().GetLogin(),
        )
}

func HandleWorkflowRunEvent(e *github.WorkflowRunEvent) string {
        if e == nil {
                return "⚠️ <b>No workflow run data</b>"
        }

        workflow := e.GetWorkflow().GetName()
        run := e.GetWorkflowRun()
        repo := e.GetRepo().GetFullName()
        repoURL := e.GetRepo().GetHTMLURL()
        sender := e.GetSender().GetLogin()

        status := run.GetStatus()
        conclusion := run.GetConclusion()

        var statusEmoji, statusText string

        switch {
        case status == "completed" && conclusion == "success":
                statusEmoji = "✅"
                statusText = "Success"
        case status == "completed" && conclusion == "failure":
                statusEmoji = "❌"
                statusText = "Failed"
        case status == "completed" && conclusion == "cancelled":
                statusEmoji = "⛔"
                statusText = "Cancelled"
        case status == "completed":
                statusEmoji = "✔️"
                statusText = "Completed"
        case status == "in_progress":
                statusEmoji = "🔄"
                statusText = "Running"
        case status == "queued":
                statusEmoji = "⏳"
                statusText = "Queued"
        default:
                statusEmoji = "⚠️"
                statusText = strings.Title(status)
        }

        // Handle Dependabot specifically
        if strings.Contains(workflow, "Dependabot") {
                statusEmoji = "🔗"
        }

        msg := fmt.Sprintf(
                "%s <b>%s</b> workflow\n"+
                        "<b>Status:</b> %s\n"+
                        "<b>Repo:</b> <a href='%s'>%s</a>\n"+
                        "<b>By:</b> %s\n",
                statusEmoji,
                formatText(workflow, 0),
                statusText,
                safeGetURL(repoURL),
                formatText(repo, 0),
                formatText(sender, 0),
        )

        if runURL := run.GetHTMLURL(); runURL != "" {
                msg += fmt.Sprintf("<a href='%s'>View Workflow Run</a>", safeGetURL(runURL))
        }

        return msg
}

func HandleWorkflowJobEvent(e *github.WorkflowJobEvent) string {
        if e == nil {
                return "⚠️ <b>No workflow job data</b>"
        }

        job := e.GetWorkflowJob()
        if job == nil {
                return "⚠️ <b>Invalid workflow job</b>"
        }

        status := job.GetStatus()
        conclusion := job.GetConclusion()

        var statusEmoji, statusText string

        switch {
        case status == "completed" && conclusion == "success":
                statusEmoji = "✅"
                statusText = "Success"
        case status == "completed" && conclusion == "failure":
                statusEmoji = "❌"
                statusText = "Failed"
        case status == "completed" && conclusion == "cancelled":
                statusEmoji = "⛔"
                statusText = "Cancelled"
        case status == "completed":
                statusEmoji = "✔️"
                statusText = "Completed"
        case status == "in_progress":
                statusEmoji = "🔄"
                statusText = "Running"
        case status == "queued":
                statusEmoji = "⏳"
                statusText = "Queued"
        default:
                statusEmoji = "⚠️"
                statusText = strings.Title(status)
        }

        // Handle Dependabot specifically
        if strings.Contains(job.GetName(), "Dependabot") {
                statusEmoji = "🔗"
        }

        msg := fmt.Sprintf(
                "%s <b>Workflow Job: %s</b>\n"+
                        "<b>Name:</b> %s\n"+
                        "<b>Status:</b> %s\n"+
                        "<b>Repo:</b> %s\n",
                statusEmoji,
                statusText,
                formatText(job.GetName(), 0),
                statusText,
                formatText(e.GetRepo().GetFullName(), 0),
        )

        if !job.GetStartedAt().IsZero() {
                msg += fmt.Sprintf("<b>Started:</b> %s\n", job.GetStartedAt().Format("2006-01-02 15:04"))
        }
        if !job.GetCompletedAt().IsZero() {
                msg += fmt.Sprintf("<b>Completed:</b> %s\n", job.GetCompletedAt().Format("2006-01-02 15:04"))
        }

        if runner := job.GetRunnerName(); runner != "" {
                msg += fmt.Sprintf("<b>Runner:</b> %s\n", formatText(runner, 0))
        }

        msg += fmt.Sprintf("<b>By:</b> %s\n", formatText(e.GetSender().GetLogin(), 0))

        if jobURL := job.GetHTMLURL(); jobURL != "" {
                msg += fmt.Sprintf("<a href='%s'>View Job Details</a>", safeGetURL(jobURL))
        }

        return msg
}

func HandleWorkflowDispatchEvent(e *github.WorkflowDispatchEvent) string {
        repo := e.GetRepo().GetFullName()
        workflow := e.GetWorkflow()
        if workflow == "" {
                workflow = "Unnamed Workflow"
        }

        inputs := "No inputs"
        if e.Inputs != nil {
                var inputsMap map[string]interface{}
                if err := json.Unmarshal(e.Inputs, &inputsMap); err == nil && len(inputsMap) > 0 {
                        var inputPairs []string
                        for k, v := range inputsMap {
                                inputPairs = append(inputPairs, fmt.Sprintf("%s: %v", k, v))
                        }
                        inputs = strings.Join(inputPairs, ", ")
                }
        }

        return fmt.Sprintf(
                "🚀 <b>%s</b> manually triggered\n"+
                        "<b>Repo:</b> %s\n"+
                        "<b>Branch:</b> %s\n"+
                        "<b>Inputs:</b> %s\n"+
                        "<b>By:</b> %s",
                formatText(workflow, 0),
                repo,
                e.GetRef(),
                formatText(inputs, 100),
                e.GetSender().GetLogin(),
        )
}

func HandleTeamAddEvent(e *github.TeamAddEvent) string {
        return fmt.Sprintf(
                "👥 <b>Team added</b>\n"+
                        "<b>Team:</b> %s\n"+
                        "<b>Repo:</b> %s\n"+
                        "<b>Org:</b> %s\n"+
                        "<b>By:</b> %s",
                formatText(e.GetTeam().GetName(), 0),
                formatText(e.GetRepo().GetFullName(), 0),
                formatText(e.GetOrg().GetLogin(), 0),
                formatText(e.GetSender().GetLogin(), 0),
        )
}

func HandleTeamEvent(e *github.TeamEvent) string {
        action := e.GetAction()
        team := e.GetTeam().GetName()
        org := e.GetOrg().GetLogin()
        sender := e.GetSender().GetLogin()

        actionInfo := map[string]struct {
                emoji string
                verb  string
        }{
                "created": {"🆕", "created"},
                "edited":  {"✏️", "modified"},
                "deleted": {"🗑️", "deleted"},
        }[action]

        if actionInfo.emoji == "" {
                actionInfo = struct {
                        emoji string
                        verb  string
                }{"⚙️", action}
        }

        return fmt.Sprintf(
                "%s <b>Team %s</b>\n"+
                        "<b>Name:</b> %s\n"+
                        "<b>Org:</b> %s\n"+
                        "<b>By:</b> %s",
                actionInfo.emoji,
                actionInfo.verb,
                formatText(team, 0),
                formatText(org, 0),
                formatText(sender, 0),
        )
}

func HandleStarEvent(e *github.StarEvent) string {
        action := e.GetAction()
        user := e.GetSender().GetLogin()
        repo := e.GetRepo().GetFullName()
        repoURL := e.GetRepo().GetHTMLURL()

        var emoji, actionText string
        switch action {
        case "created":
                emoji = "⭐"
                actionText = "starred"
        case "deleted":
                emoji = "❌"
                actionText = "unstarred"
        default:
                emoji = "⚠️"
                actionText = "performed unknown action on"
        }

        return fmt.Sprintf(
                "%s <a href='https://github.com/%s'>%s</a> %s <a href='%s'>%s</a>",
                emoji,
                user,
                user,
                actionText,
                safeGetURL(repoURL),
                repo,
        )
}

func HandleRepositoryDispatchEvent(e *github.RepositoryDispatchEvent) string {
        repo := e.GetRepo().GetFullName()
        sender := e.GetSender().GetLogin()
        action := e.GetAction()
        branch := e.Branch
        if branch == nil {
                branch = e.Repo.MasterBranch
        }

        var payloadStr string
        if e.ClientPayload != nil {
                var payload map[string]interface{}
                if err := json.Unmarshal(e.ClientPayload, &payload); err == nil {
                        if len(payload) > 0 {
                                payloadBytes, _ := json.Marshal(payload)
                                payloadStr = fmt.Sprintf("\n<b>Payload:</b> <pre>%s</pre>", formatText(string(payloadBytes), 300))
                        }
                }
        }

        return fmt.Sprintf(
                "🚀 <b>Repository Dispatch</b>\n"+
                        "<b>Repo:</b> %s\n"+
                        "<b>Action:</b> %s\n"+
                        "<b>Branch:</b> %s\n"+
                        "<b>By:</b> %s%s",
                repo,
                action,
                branchOrDefault(branch),
                sender,
                payloadStr,
        )
}

func HandlePullRequestReviewCommentEvent(e *github.PullRequestReviewCommentEvent) string {
        action := e.GetAction()
        repo := e.GetRepo().GetFullName()
        comment := e.GetComment()
        pr := e.GetPullRequest()

        actionEmoji := map[string]string{
                "created": "💬",
                "edited":  "✏️",
                "deleted": "🗑️",
        }[action]
        if actionEmoji == "" {
                actionEmoji = "⚠️"
        }

        return fmt.Sprintf(
                "%s <b>PR Review Comment %s</b>\n"+
                        "<b>Repo:</b> %s\n"+
                        "<b>PR:</b> <a href='%s'>#%d %s</a>\n"+
                        "<b>Comment:</b> %s\n"+
                        "<a href='%s'>View Comment</a>",
                actionEmoji,
                action,
                repo,
                safeGetURL(pr.GetHTMLURL()),
                pr.GetNumber(),
                formatText(pr.GetTitle(), 100),
                formatText(comment.GetBody(), 120),
                safeGetURL(comment.GetHTMLURL()),
        )
}

func HandlePullRequestReviewEvent(e *github.PullRequestReviewEvent) string {
        action := e.GetAction()
        review := e.GetReview()
        pr := e.GetPullRequest()

        stateEmoji := map[string]string{
                "approved":          "✅",
                "changes_requested": "✏️",
                "commented":         "💬",
                "dismissed":         "❌",
        }[review.GetState()]

        if stateEmoji == "" {
                stateEmoji = "🔍"
        }

        return fmt.Sprintf(
                "%s <b>PR Review %s</b>\n"+
                        "<b>Repo:</b> %s\n"+
                        "<b>PR:</b> <a href='%s'>#%d %s</a>\n"+
                        "<b>State:</b> %s\n"+
                        "<b>By:</b> %s\n"+
                        "<a href='%s'>View Review</a>",
                stateEmoji,
                action,
                e.GetRepo().GetFullName(),
                safeGetURL(pr.GetHTMLURL()),
                pr.GetNumber(),
                formatText(pr.GetTitle(), 100),
                review.GetState(),
                e.GetSender().GetLogin(),
                safeGetURL(review.GetHTMLURL()),
        )
}

func HandlePingEvent(e *github.PingEvent) string {
        msg := "🏓 <b>Webhook Ping Received</b>\n"

        if e.Zen != nil {
                msg += fmt.Sprintf("🧘 <i>%s</i>\n", formatText(*e.Zen, 0))
        }

        if e.Repo != nil {
                msg += fmt.Sprintf(
                        "📦 <a href='https://github.com/%s'>%s</a>\n",
                        *e.Repo.FullName,
                        formatText(*e.Repo.Name, 0),
                )
        }

        if e.Sender != nil {
                msg += fmt.Sprintf("👤 By: %s\n", formatText(*e.Sender.Login, 0))
        }

        if e.Org != nil {
                msg += fmt.Sprintf("🏢 Org: %s", formatText(*e.Org.Login, 0))
        }

        return msg
}

func HandlePageBuildEvent(e *github.PageBuildEvent) string {
        msg := "🏗️ <b>GitHub Pages Build</b>\n"

        if e.Build != nil {
                status := "unknown"
                if e.Build.Status != nil {
                        status = *e.Build.Status
                }
                msg += fmt.Sprintf("<b>Status:</b> %s\n", status)

                if e.Build.Error != nil {
                        msg += fmt.Sprintf("<b>Error:</b> %v\n", formatText(*e.Build.Error.Message, 0))
                }
        }

        if e.Repo != nil {
                msg += fmt.Sprintf(
                        "📦 <a href='https://github.com/%s'>%s</a>\n",
                        *e.Repo.FullName,
                        formatText(*e.Repo.Name, 0),
                )
        }

        if e.Sender != nil {
                msg += fmt.Sprintf("👤 By: %s", formatText(*e.Sender.Login, 0))
        }

        return msg
}

func HandlePackageEvent(e *github.PackageEvent) string {
        msg := "📦 <b>Package Event</b>\n"

        if e.Package != nil && e.Package.Name != nil {
                msg += fmt.Sprintf("<b>Package:</b> %s\n", formatText(*e.Package.Name, 0))
        }

        if e.Repo != nil && e.Repo.Name != nil {
                msg += fmt.Sprintf(
                        "<b>Repo:</b> <a href='https://github.com/%s'>%s</a>\n",
                        *e.Repo.FullName,
                        formatText(*e.Repo.Name, 0),
                )
        }

        if e.Sender != nil && e.Sender.Login != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", formatText(*e.Sender.Login, 0))
        }

        return msg
}

func HandleOrgBlockEvent(e *github.OrgBlockEvent) string {
        msg := "🚫 <b>Organization Block</b>\n"

        if user := e.GetBlockedUser(); user != nil {
                msg += fmt.Sprintf("<b>Blocked:</b> %s\n", user.GetLogin())
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleOrganizationEvent(e *github.OrganizationEvent) string {
        action := e.GetAction()
        sender := e.GetSender()

        msg := fmt.Sprintf("🏢 <b>Organization Event</b>\n<b>Action:</b> %s", action)

        if sender != nil {
                msg += fmt.Sprintf("\n<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleMilestoneEvent(e *github.MilestoneEvent) string {
        milestone := e.GetMilestone()
        action := e.GetAction()

        msg := fmt.Sprintf("🏁 <b>Milestone %s</b>\n", action)

        if milestone != nil {
                msg += fmt.Sprintf("<b>Title:</b> %s\n", formatText(milestone.GetTitle(), 0))
                if desc := milestone.GetDescription(); desc != "" {
                        msg += fmt.Sprintf("<b>Description:</b> %s\n", formatText(desc, 100))
                }
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleMetaEvent(e *github.MetaEvent) string {
        msg := "⚙️ <b>Meta Event</b>\n"

        if id := e.GetHookID(); id != 0 {
                msg += fmt.Sprintf("<b>Hook ID:</b> %d\n", id)
        }

        if repo := e.GetRepo(); repo != nil {
                msg += fmt.Sprintf("<b>Repo:</b> %s\n", repo.GetName())
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s\n", sender.GetLogin())
        }

        if org := e.GetOrg(); org != nil {
                msg += fmt.Sprintf("<b>Org:</b> %s\n", org.GetLogin())
        }

        if install := e.GetInstallation(); install != nil {
                msg += fmt.Sprintf("<b>Install ID:</b> %d", install.GetID())
        }

        return msg
}

func HandleMembershipEvent(e *github.MembershipEvent) string {
        if e == nil {
                return "🚫 <b>No membership event data</b>"
        }

        msg := fmt.Sprintf("👥 <b>Membership %s</b>\n", e.GetAction())

        if scope := e.GetScope(); scope != "" {
                msg += fmt.Sprintf("<b>Scope:</b> %s\n", scope)
        }

        if member := e.GetMember(); member != nil {
                msg += fmt.Sprintf("<b>Member:</b> %s\n", member.GetLogin())
        }

        if team := e.GetTeam(); team != nil {
                msg += fmt.Sprintf("<b>Team:</b> %s\n", team.GetName())
                if desc := team.GetDescription(); desc != "" {
                        msg += fmt.Sprintf("<b>Description:</b> %s\n", formatText(desc, 100))
                }
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleDeploymentEvent(e *github.DeploymentEvent) string {
        msg := "🚀 <b>Deployment Event</b>\n"

        if deploy := e.GetDeployment(); deploy != nil {
                msg += fmt.Sprintf("<b>ID:</b> %d\n", deploy.GetID())
                if desc := deploy.GetDescription(); desc != "" {
                        msg += fmt.Sprintf("<b>Description:</b> %s\n", formatText(desc, 100))
                }
        }

        if repo := e.GetRepo(); repo != nil {
                msg += fmt.Sprintf("<b>Repo:</b> %s\n", repo.GetName())
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleLabelEvent(e *github.LabelEvent) string {
        if e == nil {
                return "🏷️ <b>No label event data</b>"
        }

        msg := fmt.Sprintf("🏷️ <b>Label %s</b>\n", e.GetAction())

        if label := e.GetLabel(); label != nil {
                msg += fmt.Sprintf("<b>Name:</b> %s\n", label.GetName())
                msg += fmt.Sprintf("<b>Color:</b> #%s\n", label.GetColor())
                if desc := label.GetDescription(); desc != "" {
                        msg += fmt.Sprintf("<b>Description:</b> %s\n", formatText(desc, 100))
                }
        }

        if changes := e.GetChanges(); changes != nil {
                if title := changes.GetTitle(); title != nil && title.GetFrom() != "" {
                        msg += fmt.Sprintf("<b>Previous Name:</b> %s\n", title.GetFrom())
                }
                if body := changes.GetBody(); body != nil && body.GetFrom() != "" {
                        msg += fmt.Sprintf("<b>Previous Desc:</b> %s\n", body.GetFrom())
                }
        }

        return msg
}

func HandleMarketplacePurchaseEvent(e *github.MarketplacePurchaseEvent) string {
        if e == nil {
                return "🛒 <b>No marketplace data</b>"
        }

        msg := fmt.Sprintf("🛒 <b>Marketplace %s</b>\n", e.GetAction())

        if purchase := e.GetMarketplacePurchase(); purchase != nil {
                if plan := purchase.GetPlan(); plan != nil {
                        msg += fmt.Sprintf("<b>Plan:</b> %s\n", plan.GetName())
                }
                msg += fmt.Sprintf("<b>Billing:</b> %s\n", purchase.GetBillingCycle())
                msg += fmt.Sprintf("<b>Units:</b> %d\n", purchase.GetUnitCount())
                if nextBill := purchase.GetNextBillingDate(); !nextBill.IsZero() {
                        msg += fmt.Sprintf("<b>Next Bill:</b> %s\n", nextBill.Format("2006-01-02"))
                }

                if account := purchase.GetAccount(); account != nil {
                        msg += fmt.Sprintf("<b>Account:</b> %s (%s)\n",
                                account.GetLogin(),
                                account.GetType())
                }
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleGollumEvent(e *github.GollumEvent) string {
        if e == nil {
                return "📚 <b>No wiki update data available</b>"
        }

        var msg strings.Builder
        msg.WriteString("📚 <b>Wiki Update</b>\n")
        if repo := e.GetRepo(); repo != nil {
                msg.WriteString(fmt.Sprintf("<b>Repository:</b> <a href=\"%s\">%s</a>\n",
                        safeGetURL(repo.GetHTMLURL()),
                        repo.GetFullName()))
        }

        if org := e.GetOrg(); org != nil {
                msg.WriteString(fmt.Sprintf("<b>Organization:</b> %s\n", org.GetLogin()))
        }

        if sender := e.GetSender(); sender != nil {
                msg.WriteString(fmt.Sprintf("<b>Edited by:</b> %s\n", sender.GetLogin()))
        }

        if e.Pages != nil && len(e.Pages) > 0 {
                msg.WriteString("\n<b>Page Changes:</b>\n")
                for _, page := range e.Pages {
                        if page == nil {
                                continue
                        }
                        action := "unknown"
                        if page.Action != nil {
                                action = *page.Action
                        }
                        emoji := getActionEmoji(action)
                        pageTitle := ""
                        if page.Title != nil {
                                pageTitle = *page.Title
                        } else if page.PageName != nil {
                                pageTitle = *page.PageName
                        }

                        if pageTitle != "" {
                                msg.WriteString(fmt.Sprintf("%s <b>%s</b> (%s)\n",
                                        emoji,
                                        formatText(pageTitle, 0),
                                        action))
                        }
                        if page.Summary != nil && *page.Summary != "" {
                                msg.WriteString(fmt.Sprintf("<i>Summary:</i> %s\n", formatText(*page.Summary, 100)))
                        }

                        if page.SHA != nil && *page.SHA != "" {
                                msg.WriteString(fmt.Sprintf("<i>Revision:</i> %s\n", (*page.SHA)[:7]))
                        }
                        if page.HTMLURL != nil && *page.HTMLURL != "" {
                                msg.WriteString(fmt.Sprintf("<a href=\"%s\">View Page</a>\n", safeGetURL(*page.HTMLURL)))
                        }

                        msg.WriteString("\n")
                }
        }

        return msg.String()
}

func getActionEmoji(action string) string {
        switch action {
        case "created":
                return "🆕"
        case "edited":
                return "✏️"
        case "deleted":
                return "🗑️"
        default:
                return "🔹"
        }
}

func HandleDeployKeyEvent(e *github.DeployKeyEvent) string {
        if e == nil {
                return "🔑 <b>No deploy key data</b>"
        }

        msg := fmt.Sprintf("🔑 <b>Deploy Key %s</b>\n", e.GetAction())

        if key := e.GetKey(); key != nil {
                msg += fmt.Sprintf("<b>Title:</b> %s\n", key.GetTitle())
                if url := key.GetURL(); url != "" {
                        msg += fmt.Sprintf("<a href=\"%s\">View Key</a>\n", safeGetURL(url))
                }
        }

        msg += fmt.Sprintf("<b>Repo:</b> %s\n", e.GetRepo().GetName())

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleCheckSuiteEvent(e *github.CheckSuiteEvent) string {
        if e == nil {
                return "✅ <b>No check suite data</b>"
        }

        suite := e.GetCheckSuite()
        var msg strings.Builder

        action := formatText(strings.Title(e.GetAction()), 0)
        msg.WriteString(fmt.Sprintf("✅ <b>Check Suite: %s</b>\n\n", action))

        if suite != nil {
                status := formatText(suite.GetStatus(), 0)
                msg.WriteString(fmt.Sprintf("• <b>Status:</b> %s\n", status))

                if conclusion := suite.GetConclusion(); conclusion != "" {
                        msg.WriteString(fmt.Sprintf("• <b>Result:</b> %s\n", formatText(conclusion, 0)))
                }

                if url := suite.GetURL(); url != "" {
                        msg.WriteString(fmt.Sprintf("\n<a href=\"%s\">🔗 View Details</a>\n", safeGetURL(url)))
                }
        }

        repo := formatText(e.GetRepo().GetFullName(), 0)
        msg.WriteString(fmt.Sprintf("\n<b>Repository:</b> %s\n", repo))

        if sender := e.GetSender(); sender != nil {
                username := formatText(sender.GetLogin(), 0)
                msg.WriteString(fmt.Sprintf("<b>Triggered by:</b> %s", username))
        }

        return msg.String()
}

func HandleCheckRunEvent(e *github.CheckRunEvent) string {
        if e == nil {
                return "⚙️ <b>No check run data</b>"
        }

        check := e.GetCheckRun()
        var msg strings.Builder

        action := formatText(strings.Title(e.GetAction()), 0)
        msg.WriteString(fmt.Sprintf("⚙️ <b>Check Run: %s</b>\n\n", action))

        if check != nil {
                name := formatText(check.GetName(), 0)
                status := formatText(check.GetStatus(), 0)
                msg.WriteString(fmt.Sprintf("• <b>Name:</b> %s\n", name))
                msg.WriteString(fmt.Sprintf("• <b>Status:</b> %s\n", status))

                if conclusion := check.GetConclusion(); conclusion != "" {
                        msg.WriteString(fmt.Sprintf("• <b>Result:</b> %s\n", formatText(conclusion, 0)))
                }

                if !check.GetStartedAt().IsZero() {
                        msg.WriteString(fmt.Sprintf("• <b>Started:</b> %s\n", check.GetStartedAt().Format("2006-01-02 15:04")))
                }

                if !check.GetCompletedAt().IsZero() {
                        msg.WriteString(fmt.Sprintf("• <b>Completed:</b> %s\n", check.GetCompletedAt().Format("2006-01-02 15:04")))
                }

                if url := check.GetHTMLURL(); url != "" {
                        msg.WriteString(fmt.Sprintf("\n<a href=\"%s\">🔗 View Details</a>\n", safeGetURL(url)))
                }
        }

        repo := formatText(e.GetRepo().GetFullName(), 0)
        msg.WriteString(fmt.Sprintf("\n<b>Repository:</b> %s\n", repo))

        if sender := e.GetSender(); sender != nil {
                username := formatText(sender.GetLogin(), 0)
                msg.WriteString(fmt.Sprintf("<b>Triggered by:</b> %s", username))
        }

        return msg.String()
}

func HandleDeploymentStatusEvent(e *github.DeploymentStatusEvent) string {
        if e == nil {
                return "🚦 <b>No deployment status data</b>"
        }

        status := e.GetDeploymentStatus()
        msg := fmt.Sprintf("🚦 <b>Deployment %s</b>\n", status.GetState())

        if desc := status.GetDescription(); desc != "" {
                msg += fmt.Sprintf("<b>Status:</b> %s\n", formatText(desc, 100))
        }

        msg += fmt.Sprintf("<b>Repo:</b> %s\n", e.GetRepo().GetName())

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}

func HandleSecurityAdvisoryEvent(e *github.SecurityAdvisoryEvent) string {
        if e == nil {
                return "⚠️ <b>No security advisory data</b>"
        }

        adv := e.GetSecurityAdvisory()
        msg := fmt.Sprintf("⚠️ <b>Security Advisory %s</b>\n", e.GetAction())

        if adv != nil {
                msg += fmt.Sprintf("<b>Summary:</b> %s\n", formatText(adv.GetSummary(), 100))
                if sev := adv.GetSeverity(); sev != "" {
                        msg += fmt.Sprintf("<b>Severity:</b> %s\n", sev)
                }
                if cve := adv.GetCVEID(); cve != "" {
                        msg += fmt.Sprintf("<b>CVE:</b> %s\n", cve)
                }
                if url := adv.GetURL(); url != "" {
                        msg += fmt.Sprintf("<a href=\"%s\">View Advisory</a>\n", safeGetURL(url))
                }
                if author := adv.GetAuthor(); author != nil {
                        msg += fmt.Sprintf("<b>Reported by:</b> %s\n", author.GetLogin())
                }
        }

        if repo := e.GetRepository(); repo != nil {
                msg += fmt.Sprintf("<b>Repo:</b> %s\n", repo.GetFullName())
        }

        if org := e.GetOrganization(); org != nil {
                msg += fmt.Sprintf("<b>Org:</b> %s\n", org.GetLogin())
        }

        if sender := e.GetSender(); sender != nil {
                msg += fmt.Sprintf("<b>By:</b> %s", sender.GetLogin())
        }

        return msg
}