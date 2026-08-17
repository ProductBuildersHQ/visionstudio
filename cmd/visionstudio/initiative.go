package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/spf13/cobra"
)

func parseSpecs(flags []string) map[string]string {
	if len(flags) == 0 {
		return nil
	}
	specs := make(map[string]string, len(flags))
	for _, f := range flags {
		k, v, ok := strings.Cut(f, "=")
		if ok {
			specs[k] = v
		}
	}
	return specs
}

func initiativeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initiative",
		Aliases: []string{"init"},
		Short:   "Manage initiatives",
	}
	cmd.AddCommand(
		initiativeCreateCmd(),
		initiativeListCmd(),
		initiativeGetCmd(),
		initiativeUpdateCmd(),
		initiativeTransitionCmd(),
		initiativeSweepCmd(),
		initiativeDepCmd(),
		initiativeHideCmd(),
		initiativeShowCmd(),
	)
	return cmd
}

// setInitiativeHidden loads an initiative, sets its hidden flag, and persists it.
func setInitiativeHidden(cmd *cobra.Command, id string, hidden bool) error {
	svc, cleanup, err := connectService(cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	init, err := svc.Store.GetInitiative(cmd.Context(), id)
	if err != nil {
		return err
	}
	if init.Hidden == hidden {
		state := "shown"
		if hidden {
			state = "hidden"
		}
		cmd.Printf("Initiative %s is already %s\n", init.ID, state)
		return nil
	}
	init.Hidden = hidden
	init.UpdatedAt = time.Now()
	if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
		return err
	}
	verb := "shown in"
	if hidden {
		verb = "hidden from"
	}
	cmd.Printf("Initiative %s %s the dashboard\n", init.ID, verb)
	return nil
}

func initiativeHideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hide <initiative-id>",
		Short: "Hide an initiative from the dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setInitiativeHidden(cmd, args[0], true)
		},
	}
}

func initiativeShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <initiative-id>",
		Short: "Show a previously hidden initiative on the dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setInitiativeHidden(cmd, args[0], false)
		},
	}
}

func initiativeCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new initiative",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			org, _ := cmd.Flags().GetString("org")
			title, _ := cmd.Flags().GetString("title")
			desc, _ := cmd.Flags().GetString("description")
			priority, _ := cmd.Flags().GetString("priority")
			initType, _ := cmd.Flags().GetString("type")
			homeRepo, _ := cmd.Flags().GetString("home-repo")
			workspace, _ := cmd.Flags().GetString("workspace")
			program, _ := cmd.Flags().GetString("program")
			specFlags, _ := cmd.Flags().GetStringSlice("spec")
			workflowID, _ := cmd.Flags().GetString("workflow")

			if id == "" || title == "" {
				return fmt.Errorf("--id and --title are required")
			}

			// Workflow is required; fall back to config default if not specified
			if workflowID == "" {
				if cfg, err := cliconfig.Load(); err == nil && cfg.Defaults.Workflow != "" {
					workflowID = cfg.Defaults.Workflow
				} else {
					return fmt.Errorf("--workflow is required (or set defaults.workflow in ~/.productbuildershq/visionstudio/config.json)")
				}
			}
			if _, err := specworkflow.DefaultLoader().Load(workflowID); err != nil {
				return fmt.Errorf("unknown workflow %q (see 'visionstudio workflow list'): %w", workflowID, err)
			}
			if org == "" {
				org = "default"
			}

			specs := parseSpecs(specFlags)

			init, err := svc.CreateInitiative(cmd.Context(), id, org, title, desc, priority, initType, workflowID)
			if err != nil {
				return err
			}

			if homeRepo != "" || workspace != "" || program != "" || len(specs) > 0 {
				init.HomeRepo = homeRepo
				init.Workspace = workspace
				init.ProgramID = program
				init.Specs = specs
				if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
					return err
				}
			}

			cmd.Printf("Created initiative: %s (%s)\n", init.ID, init.Status)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Initiative ID (e.g. INIT-MYPROJECT-001) (required)")
	cmd.Flags().String("org", "", "Organization (default: 'default')")
	cmd.Flags().String("title", "", "Initiative title (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("priority", "", "Priority (high, medium, low)")
	cmd.Flags().String("type", "", "Initiative type (feature, maintenance, migration, compliance, refactor); default: feature")
	cmd.Flags().String("home-repo", "", "Home repository ID (where specs live)")
	cmd.Flags().String("workspace", "", "Workspace identifier (e.g. tmux session name)")
	cmd.Flags().String("program", "", "Program ID (e.g. PROG-DELIVERY)")
	cmd.Flags().StringSlice("spec", nil, "Spec reference as key=path (repeatable, e.g. --spec prd=docs/specs/PRD.md)")
	cmd.Flags().String("workflow", "", "Spec workflow ID (e.g. pbhq-lite, aws-one-way-door, big-tech-essentials)")
	return cmd
}

func initiativeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all initiatives",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			inits, err := svc.ListInitiatives(cmd.Context())
			if err != nil {
				return err
			}

			repoFilter, _ := cmd.Flags().GetString("repo")
			statusFilter, _ := cmd.Flags().GetString("status")
			programFilter, _ := cmd.Flags().GetString("program")

			if repoFilter != "" {
				repoFilter, err = resolveRepoID(cmd.Context(), svc, repoFilter)
				if err != nil {
					return err
				}
			}

			if repoFilter != "" || statusFilter != "" || programFilter != "" {
				filtered := make([]*store.Initiative, 0, len(inits))
				for _, i := range inits {
					if repoFilter != "" && i.HomeRepo != repoFilter {
						continue
					}
					if statusFilter != "" && i.Status != statusFilter {
						continue
					}
					if programFilter != "" && i.ProgramID != programFilter {
						continue
					}
					filtered = append(filtered, i)
				}
				inits = filtered
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(inits)
			case "text":
				// falls through to the tabwriter rendering below
			default:
				return fmt.Errorf("unknown format: %s (use text or json)", format)
			}

			if len(inits) == 0 {
				cmd.Println("No initiatives found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPROGRAM\tWORKFLOW\tWORKSPACE\tHIDDEN")
			for _, i := range inits {
				ws := i.Workspace
				if ws == "" {
					ws = "-"
				}
				prog := i.ProgramID
				if prog == "" {
					prog = "-"
				}
				wf := i.WorkflowID
				if wf == "" {
					wf = "-"
				}
				hidden := "no"
				if i.Hidden {
					hidden = "yes"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", i.ID, i.Title, i.Status, prog, wf, ws, hidden)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("format", "text", "Output format: text or json")
	cmd.Flags().String("repo", "", "Filter by home repository (short name, org/name, or full ID)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("program", "", "Filter by program ID")
	return cmd
}

func initiativeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <initiative-id>",
		Short: "Show initiative details with phase status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			detail, err := svc.GetInitiativeDetail(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			init := detail.Initiative
			cmd.Printf("Initiative: %s\n", init.ID)
			cmd.Printf("Title:      %s\n", init.Title)
			cmd.Printf("Status:     %s\n", init.Status)
			if init.Description != "" {
				cmd.Printf("Desc:       %s\n", init.Description)
			}
			if init.Priority != "" {
				cmd.Printf("Priority:   %s\n", init.Priority)
			}
			if init.HomeRepo != "" {
				cmd.Printf("Home repo:  %s\n", init.HomeRepo)
			}
			if init.Workspace != "" {
				cmd.Printf("Workspace:  %s\n", init.Workspace)
			}
			if init.ProgramID != "" {
				cmd.Printf("Program:    %s\n", init.ProgramID)
			}
			if init.WorkflowID != "" {
				cmd.Printf("Workflow:   %s\n", init.WorkflowID)
			} else {
				cmd.Printf("Workflow:   %s (default for type %s)\n",
					specworkflow.DefaultWorkflowForType(init.InitType), init.InitType)
			}
			if init.Hidden {
				cmd.Printf("Hidden:     true\n")
			}
			if len(init.Specs) > 0 {
				cmd.Println("Specs:")
				for k, v := range init.Specs {
					cmd.Printf("  %s: %s\n", k, v)
				}
			}
			cmd.Printf("Created:    %s\n", init.CreatedAt.Format("2006-01-02 15:04"))
			if init.PlannedAt != nil {
				cmd.Printf("Planned:    %s\n", init.PlannedAt.Format("2006-01-02 15:04"))
			}
			if init.ExecutingAt != nil {
				cmd.Printf("Executing:  %s\n", init.ExecutingAt.Format("2006-01-02 15:04"))
			}
			if init.DeliveryCompleteAt != nil {
				cmd.Printf("Delivered:  %s\n", init.DeliveryCompleteAt.Format("2006-01-02 15:04"))
			}
			if init.ReleasedAt != nil {
				cmd.Printf("Released:   %s\n", init.ReleasedAt.Format("2006-01-02 15:04"))
			}
			if init.ClosedAt != nil {
				cmd.Printf("Closed:     %s\n", init.ClosedAt.Format("2006-01-02 15:04"))
			}
			if len(detail.Releases) > 0 {
				cmd.Println("Releases:")
				for _, r := range detail.Releases {
					cmd.Printf("  %s@%s (%s)\n", r.RepositoryID, r.Tag, r.ReleasedAt.Format("2006-01-02"))
				}
			}

			if len(detail.Phases) > 0 {
				var totalRMIs, completedRMIs, requiredTotal, requiredCompleted int
				for _, p := range detail.Phases {
					for _, r := range p.RMIs {
						totalRMIs++
						if r.Status == "completed" {
							completedRMIs++
						}
						if r.Required {
							requiredTotal++
							if r.Status == "completed" {
								requiredCompleted++
							}
						}
					}
				}
				cmd.Printf("\nProgress:   %d/%d RMIs completed", completedRMIs, totalRMIs)
				if requiredTotal != totalRMIs {
					cmd.Printf(" (%d/%d required)", requiredCompleted, requiredTotal)
				}
				cmd.Println()

				cmd.Println("\nPhases:")
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "  #\tTITLE\tTHEME\tSTATUS\tRMIs")
				for _, p := range detail.Phases {
					theme := p.Phase.Theme
					if theme == "" {
						theme = "-"
					}
					_, _ = fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%d\n",
						p.Phase.SequenceNumber, p.Phase.Title, theme, p.Status, len(p.RMIs))
				}
				if err := w.Flush(); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func initiativeUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <initiative-id>",
		Short: "Update initiative fields (workspace, home-repo, description, priority)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			init, err := svc.Store.GetInitiative(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("workspace") {
				init.Workspace, _ = cmd.Flags().GetString("workspace")
			}
			if cmd.Flags().Changed("home-repo") {
				init.HomeRepo, _ = cmd.Flags().GetString("home-repo")
			}
			if cmd.Flags().Changed("description") {
				init.Description, _ = cmd.Flags().GetString("description")
			}
			if cmd.Flags().Changed("priority") {
				init.Priority, _ = cmd.Flags().GetString("priority")
			}
			if cmd.Flags().Changed("program") {
				init.ProgramID, _ = cmd.Flags().GetString("program")
			}
			if cmd.Flags().Changed("visibility") {
				vis, _ := cmd.Flags().GetString("visibility")
				if vis != "internal" && vis != "public" {
					return fmt.Errorf("--visibility must be internal or public")
				}
				init.Visibility = vis
			}
			if cmd.Flags().Changed("workflow") {
				workflowID, _ := cmd.Flags().GetString("workflow")
				if workflowID != "" {
					if _, err := specworkflow.DefaultLoader().Load(workflowID); err != nil {
						return fmt.Errorf("unknown workflow %q (see 'visionstudio workflow list'): %w", workflowID, err)
					}
				}
				init.WorkflowID = workflowID
			}

			if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
				return err
			}
			// Keep the workflow-selection record (read by synthesis and
			// evaluation) in step with the initiative's workflow edge.
			if cmd.Flags().Changed("workflow") && init.WorkflowID != "" {
				if err := svc.Store.SelectWorkflowForInitiative(cmd.Context(), init.ID, init.WorkflowID); err != nil {
					return fmt.Errorf("update workflow selection: %w", err)
				}
			}

			cmd.Printf("Updated %s\n", init.ID)
			if init.Workspace != "" {
				cmd.Printf("Workspace: %s\n", init.Workspace)
			}
			if cmd.Flags().Changed("workflow") {
				cmd.Printf("Workflow: %s\n", init.WorkflowID)
			}
			return nil
		},
	}
	cmd.Flags().String("workspace", "", "Workspace identifier (e.g. tmux session name)")
	cmd.Flags().String("home-repo", "", "Home repository ID")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("priority", "", "Priority (high, medium, low)")
	cmd.Flags().String("program", "", "Program ID (e.g. PROG-DELIVERY)")
	cmd.Flags().String("workflow", "", "Spec workflow ID (e.g. pbhq-lite, aws-one-way-door, aws-two-way-door); empty string clears the override")
	cmd.Flags().String("visibility", "", "Projection visibility: internal (default) or public — the flag-flip is the publication approval")
	return cmd
}

func initiativeTransitionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transition <initiative-id> <status>",
		Short: "Transition initiative to a new lifecycle status",
		Long: `Valid transitions:
  Forward:   proposed -> planned -> executing -> delivery_complete -> releasing -> released -> closed
  Backwards: any status may reopen to any earlier pipeline status as scope
             evolves (e.g. delivery_complete -> executing when new phases land);
             reopening clears the lifecycle timestamps of the stages it undoes
  Cancel:    any pre-release status -> cancelled; cancelled may reopen to any
             pre-release status (never to released or closed)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			init, err := svc.TransitionInitiative(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			cmd.Printf("Transitioned %s to %s\n", init.ID, init.Status)
			return nil
		},
	}
}

// sweepRepoCheck is the git-state verdict for one repository referenced by
// a sweep candidate's RMIs.
type sweepRepoCheck struct {
	RepositoryID string `json:"repository_id"`
	LocalPath    string `json:"local_path,omitempty"`
	State        string `json:"state"`
	Detail       string `json:"detail"`
}

// sweepCandidate is a non-terminal initiative whose RMIs are all completed.
type sweepCandidate struct {
	InitiativeID string           `json:"initiative_id"`
	Title        string           `json:"title"`
	Status       string           `json:"status"`
	RMICount     int              `json:"rmi_count"`
	Repos        []sweepRepoCheck `json:"repos"`
	NeedsReview  bool             `json:"needs_review"`
}

func initiativeSweepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "Find initiatives whose RMI completion has outrun their recorded status",
		Long: `Lists non-terminal initiatives (proposed/planned/executing) where every RMI
is completed. For each candidate, resolves every distinct repository referenced
by its RMIs -- not just the initiative's home repo -- and reports local git
state: clean/dirty working tree, ahead/behind the cached remote-tracking ref
(no network fetch), or not found/not registered locally. Same best-effort,
report-only posture as 'registry doctor'.

sweep never calls transition or release record itself, and it cannot verify
that a completed RMI's shipped code actually matches its written description
-- that judgment call stays with whoever reviews the report.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			inits, err := svc.ListInitiatives(cmd.Context())
			if err != nil {
				return err
			}

			repos, err := svc.ListRepositories(cmd.Context())
			if err != nil {
				return err
			}
			repoByID := make(map[string]*store.Repository, len(repos))
			for _, r := range repos {
				repoByID[r.ID] = r
			}

			nonTerminal := map[string]bool{
				initiative.StatusProposed:  true,
				initiative.StatusPlanned:   true,
				initiative.StatusExecuting: true,
			}

			var candidates []sweepCandidate
			for _, init := range inits {
				if !nonTerminal[init.Status] {
					continue
				}
				rmis, err := svc.ListRMIs(cmd.Context(), init.ID)
				if err != nil {
					return err
				}
				if len(rmis) == 0 {
					continue
				}
				allDone := true
				repoIDSet := map[string]bool{}
				for _, r := range rmis {
					if r.Status != "completed" {
						allDone = false
						break
					}
					if r.RepositoryID != "" {
						repoIDSet[r.RepositoryID] = true
					}
				}
				if !allDone {
					continue
				}
				if len(repoIDSet) == 0 && init.HomeRepo != "" {
					repoIDSet[init.HomeRepo] = true
				}

				repoIDs := make([]string, 0, len(repoIDSet))
				for id := range repoIDSet {
					repoIDs = append(repoIDs, id)
				}
				sort.Strings(repoIDs)

				needsReview := false
				checks := make([]sweepRepoCheck, 0, len(repoIDs))
				for _, id := range repoIDs {
					check := sweepRepoGitState(id, repoByID[id])
					if check.State != "CLEAN" {
						needsReview = true
					}
					checks = append(checks, check)
				}

				candidates = append(candidates, sweepCandidate{
					InitiativeID: init.ID,
					Title:        init.Title,
					Status:       init.Status,
					RMICount:     len(rmis),
					Repos:        checks,
					NeedsReview:  needsReview,
				})
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(candidates)
			}

			if len(candidates) == 0 {
				cmd.Println("No candidates -- every non-terminal initiative has at least one incomplete RMI.")
				return nil
			}

			for _, c := range candidates {
				marker := "ready"
				if c.NeedsReview {
					marker = "needs review"
				}
				cmd.Printf("%s [%s -> %s] %s (%d RMIs, all completed)\n", c.InitiativeID, c.Status, marker, c.Title, c.RMICount)
				for _, r := range c.Repos {
					cmd.Printf("  %-12s %-45s %s\n", r.State, r.RepositoryID, r.Detail)
				}
				cmd.Println()
			}
			cmd.Printf("%d candidate(s). Verify each RMI's shipped work actually matches its spec before transitioning or recording a release.\n", len(candidates))
			return nil
		},
	}
	cmd.Flags().String("format", "text", "Output format: text or json")
	return cmd
}

// sweepRepoGitState reports the local git state of repo (best-effort, no
// network fetch -- reads cached remote-tracking refs only). Never fatal:
// lookup failures degrade to an informative state string.
func sweepRepoGitState(repoID string, repo *store.Repository) sweepRepoCheck {
	if repo == nil {
		return sweepRepoCheck{RepositoryID: repoID, State: "UNREGISTERED", Detail: "not in the repository registry"}
	}
	if repo.LocalPath == "" {
		return sweepRepoCheck{RepositoryID: repoID, State: "NO-PATH", Detail: "no local path registered"}
	}

	check := sweepRepoCheck{RepositoryID: repoID, LocalPath: repo.LocalPath}
	info, err := os.Stat(repo.LocalPath)
	switch {
	case err != nil:
		check.State = "MISSING"
		check.Detail = fmt.Sprintf("%s does not exist", repo.LocalPath)
		return check
	case !info.IsDir():
		check.State = "NOT-A-DIR"
		check.Detail = fmt.Sprintf("%s is not a directory", repo.LocalPath)
		return check
	}
	if _, err := os.Stat(filepath.Join(repo.LocalPath, ".git")); err != nil {
		check.State = "NOT-GIT"
		check.Detail = fmt.Sprintf("%s is not a git working tree", repo.LocalPath)
		return check
	}

	dirty := sweepGitDirty(repo.LocalPath)
	ahead, behind, hasUpstream := sweepGitAheadBehind(repo.LocalPath)

	var notes []string
	if dirty {
		notes = append(notes, "uncommitted changes")
	}
	switch {
	case !hasUpstream:
		notes = append(notes, "no upstream tracking branch configured")
	default:
		if ahead > 0 {
			notes = append(notes, fmt.Sprintf("%d commit(s) ahead of upstream (unpushed)", ahead))
		}
		if behind > 0 {
			notes = append(notes, fmt.Sprintf("%d commit(s) behind upstream", behind))
		}
	}

	switch {
	case dirty:
		check.State = "DIRTY"
	case hasUpstream && ahead > 0:
		check.State = "UNPUSHED"
	case !hasUpstream:
		check.State = "UNVERIFIED"
	case behind > 0:
		check.State = "BEHIND"
	default:
		check.State = "CLEAN"
	}
	if len(notes) == 0 {
		check.Detail = "clean, in sync with upstream"
	} else {
		check.Detail = strings.Join(notes, "; ")
	}
	return check
}

// sweepGitDirty reports whether path's working tree has uncommitted changes.
// Best-effort: a lookup failure is reported as not dirty (git status itself
// covers the "not a git repo" case earlier in the caller).
func sweepGitDirty(path string) bool {
	// #nosec G204 -- path is a caller-supplied local directory for a local
	// dev CLI, not untrusted network input (same posture as registryDoctorCmd).
	out, err := exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// sweepGitAheadBehind reports how many commits path's HEAD is ahead of and
// behind its upstream tracking ref, read from the local cache -- no network
// fetch. ok is false if no upstream is configured or the lookup fails.
func sweepGitAheadBehind(path string) (ahead, behind int, ok bool) {
	// #nosec G204 -- path is a caller-supplied local directory for a local
	// dev CLI, not untrusted network input (same posture as registryDoctorCmd).
	out, err := exec.Command("git", "-C", path, "rev-list", "--left-right", "--count", "HEAD...@{u}").Output()
	if err != nil {
		return 0, 0, false
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return 0, 0, false
	}
	a, errA := strconv.Atoi(fields[0])
	b, errB := strconv.Atoi(fields[1])
	if errA != nil || errB != nil {
		return 0, 0, false
	}
	return a, b, true
}

func initiativeDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dep",
		Aliases: []string{"dependency"},
		Short:   "Manage initiative dependencies",
	}
	cmd.AddCommand(initiativeDepAddCmd(), initiativeDepListCmd())
	return cmd
}

func initiativeDepAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a dependency between two initiatives",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			source, _ := cmd.Flags().GetString("source")
			target, _ := cmd.Flags().GetString("target")
			rel, _ := cmd.Flags().GetString("relationship")

			if source == "" || target == "" {
				return fmt.Errorf("--source and --target are required")
			}
			if rel == "" {
				rel = "requires"
			}

			dep := &store.InitiativeDependency{
				SourceInitiativeID: source,
				TargetInitiativeID: target,
				Relationship:       rel,
			}
			if err := svc.Store.CreateInitiativeDependency(cmd.Context(), dep); err != nil {
				return err
			}
			cmd.Printf("Added dependency: %s --%s--> %s\n", source, rel, target)
			return nil
		},
	}
	cmd.Flags().String("source", "", "Source initiative ID (required)")
	cmd.Flags().String("target", "", "Target initiative ID (required)")
	cmd.Flags().String("relationship", "requires", "Relationship type (requires, relates)")
	return cmd
}

func initiativeDepListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [initiative-id]",
		Short: "List initiative dependencies (all or for a specific initiative)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			var deps []*store.InitiativeDependency
			if len(args) == 1 {
				deps, err = svc.Store.ListInitiativeDependencies(cmd.Context(), args[0])
			} else {
				deps, err = svc.Store.ListAllInitiativeDependencies(cmd.Context())
			}
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				cmd.Println("No initiative dependencies found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "SOURCE\tRELATIONSHIP\tTARGET")
			for _, d := range deps {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", d.SourceInitiativeID, d.Relationship, d.TargetInitiativeID)
			}
			return w.Flush()
		},
	}
}
