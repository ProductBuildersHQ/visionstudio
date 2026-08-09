package main

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/report"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
)

func dashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Open a browser dashboard showing initiatives, phases, and RMIs",
		Long: `Generate and display a two-level dashboard.

The landing page shows a summary with initiative cards grouped by program,
progress bars, and RMI status distribution. Click an initiative or program
to drill down into phase/RMI detail.

By default, starts a local HTTP server that re-queries the database on
every page load, so the dashboard always shows current data.
Use --static to write a one-shot HTML file instead (summary only).
Use --unified to serve the React SPA from web/dist/ instead of Go templates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			static, _ := cmd.Flags().GetBool("static")
			unified, _ := cmd.Flags().GetBool("unified")
			port, _ := cmd.Flags().GetInt("port")

			if static {
				return runDashboardStatic(cmd)
			}
			if unified {
				return runUnifiedDashboardServer(cmd, port)
			}
			return runDashboardServer(cmd, port)
		},
	}
	cmd.Flags().Bool("static", false, "Write a static HTML file instead of running a server")
	cmd.Flags().Bool("unified", false, "Serve the unified React SPA from web/dist/")
	cmd.Flags().Int("port", 9400, "Port for the dashboard server")
	cmd.Flags().String("data-dir", "", "Path to omnidevx data directory (default: ~/.plexusone/omnidevx/data)")
	return cmd
}

func runDashboardStatic(cmd *cobra.Command) error {
	svc, cleanup, err := connectService(cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	dataDir, _ := cmd.Flags().GetString("data-dir")
	data, err := loadDashboardData(cmd.Context(), svc, dataDir)
	if err != nil {
		return err
	}

	html, err := renderSummary(data)
	if err != nil {
		return err
	}

	tmpDir := os.TempDir()
	outPath := filepath.Join(tmpDir, "vistudio-dashboard.html")
	if err := os.WriteFile(outPath, html, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	cmd.Printf("Dashboard written to %s\n", outPath)
	return openBrowser(outPath)
}

func runUnifiedDashboardServer(cmd *cobra.Command, port int) error {
	mux := http.NewServeMux()
	dataDir, _ := cmd.Flags().GetString("data-dir")

	connectSvc := func() (*service.Service, func(), error) {
		return connectService(cmd)
	}
	registerAPIRoutes(mux, connectSvc, dataDir)

	// Start periodic Dolt commit (commits every 5 minutes to prevent data loss)
	stopCommitter := startPeriodicCommitter(cmd.Context(), cmd)
	defer stopCommitter()

	// Find web/dist directory relative to the executable or working directory
	webDistPath := findWebDistPath()
	if webDistPath == "" {
		return fmt.Errorf("web/dist not found; run 'npm run build' in the web directory first")
	}

	// Serve the React SPA with fallback to index.html for client-side routing
	fs := http.FileServer(http.Dir(webDistPath))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent directory traversal
		cleanPath := filepath.Clean(r.URL.Path)
		if cleanPath == "." {
			cleanPath = "/"
		}
		fullPath := filepath.Join(webDistPath, cleanPath) //nolint:gosec // G703: path is cleaned and webDistPath is trusted
		if _, err := os.Stat(fullPath); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		// Fall back to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(webDistPath, "index.html"))
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	dashURL := fmt.Sprintf("http://%s", addr)
	cmd.Printf("Unified dashboard server running at %s (Ctrl-C to stop)\n", dashURL)
	if err := openBrowser(dashURL); err != nil {
		cmd.Printf("Open %s in your browser\n", dashURL)
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

func findWebDistPath() string {
	// Try relative to working directory
	candidates := []string{
		"web/dist",
		"../web/dist",
		"../../web/dist",
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "web/dist"),
			filepath.Join(exeDir, "../web/dist"),
			filepath.Join(exeDir, "../../web/dist"),
		)
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			indexPath := filepath.Join(absPath, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				return absPath
			}
		}
	}
	return ""
}

func runDashboardServer(cmd *cobra.Command, port int) error {
	mux := http.NewServeMux()

	dataDir, _ := cmd.Flags().GetString("data-dir")

	// Register JSON API routes
	connectSvc := func() (*service.Service, func(), error) {
		return connectService(cmd)
	}
	registerAPIRoutes(mux, connectSvc, dataDir)

	serve := func(w http.ResponseWriter, _ *http.Request, render func(*service.Service) ([]byte, error)) {
		svc, cleanup, err := connectService(cmd)
		if err != nil {
			http.Error(w, fmt.Sprintf("connect: %v", err), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		html, err := render(svc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(html); err != nil {
			return
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		showHidden := r.URL.Query().Get("show_hidden") == "1"
		serve(w, r, func(svc *service.Service) ([]byte, error) {
			data, err := loadDashboardData(r.Context(), svc, dataDir)
			if err != nil {
				return nil, err
			}
			data.ShowHidden = showHidden
			return renderSummary(data)
		})
	})

	mux.HandleFunc("/initiative/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/initiative/")
		if id == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		serve(w, r, func(svc *service.Service) ([]byte, error) {
			data, err := loadDashboardData(r.Context(), svc, dataDir)
			if err != nil {
				return nil, err
			}
			return renderInitiativeDetail(data, id)
		})
	})

	mux.HandleFunc("/program/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/program/")
		if id == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		serve(w, r, func(svc *service.Service) ([]byte, error) {
			data, err := loadDashboardData(r.Context(), svc, dataDir)
			if err != nil {
				return nil, err
			}
			return renderProgramDetail(data, id)
		})
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	dashURL := fmt.Sprintf("http://%s", addr)
	cmd.Printf("Dashboard server running at %s (Ctrl-C to stop)\n", dashURL)
	if err := openBrowser(dashURL); err != nil {
		cmd.Printf("Open %s in your browser\n", dashURL)
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// --- data model ---

type dashboardRMI struct {
	RMI       *store.RoadmapItem
	ClaimedAt string
	ClaimedBy string
	Tooltip   string
	Tokens    *report.TokenTotals // nil if no token data
}

type phaseData struct {
	Phase  *store.Phase
	RMIs   []dashboardRMI
	Tokens *report.TokenTotals // nil if no token data
}

type repoCount struct {
	Name  string
	Count int
}

type initData struct {
	Initiative    *store.Initiative
	ProgramName   string
	Phases        []phaseData
	Repos         []repoCount
	TotalRMIs     int
	CompletedRMIs int
	Tokens        *report.TokenTotals // nil if no token data
	TokenReport   *report.TokenReport // full report for model breakdown
}

type programData struct {
	ID          string
	Name        string
	Description string
	Hidden      bool
	Initiatives []initData
	Tokens      *report.TokenTotals // nil if no token data
}

type dashboardData struct {
	Initiatives   []initData
	Programs      []programData
	Standalone    []initData
	AllDeps       []*store.RMIDependency
	InitDeps      []*store.InitiativeDependency
	StatusDist    []statusCount
	TotalTokens   *report.TokenTotals // nil if no token data
	HasTokenData  bool
	TokenDataNote string // e.g., "No omnidevx data found"

	// ShowHidden, when true, reveals programs flagged Hidden on the homepage.
	// Set per-request from the ?show_hidden query param.
	ShowHidden bool
}

// VisiblePrograms returns the programs shown on the homepage: all of them when
// ShowHidden is set, otherwise only those not flagged Hidden.
func (d *dashboardData) VisiblePrograms() []programData {
	if d.ShowHidden {
		return d.Programs
	}
	visible := make([]programData, 0, len(d.Programs))
	for _, p := range d.Programs {
		if !p.Hidden {
			visible = append(visible, p)
		}
	}
	return visible
}

// HiddenProgramCount reports how many programs are flagged Hidden.
func (d *dashboardData) HiddenProgramCount() int {
	n := 0
	for _, p := range d.Programs {
		if p.Hidden {
			n++
		}
	}
	return n
}

type statusCount struct {
	Status string
	Count  int
}

func loadDashboardData(ctx context.Context, svc *service.Service, dataDir string) (*dashboardData, error) {
	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}

	allAssignments, err := svc.Store.ListAllAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	assignmentsByRMI := map[string]*store.Assignment{}
	for _, a := range allAssignments {
		existing, ok := assignmentsByRMI[a.RMIID]
		if !ok || a.CreatedAt.After(existing.CreatedAt) {
			assignmentsByRMI[a.RMIID] = a
		}
	}

	// Load token data if available
	tokenReports := make(map[string]*report.TokenReport) // by initiative ID
	var totalTokens *report.TokenTotals
	hasTokenData := false
	tokenDataNote := ""

	tokenSource, err := tokens.NewJSONLSource(dataDir)
	if err == nil {
		// Check if data directory exists
		eventsDir := filepath.Join(tokenSource.Dir, "events")
		if _, statErr := os.Stat(eventsDir); statErr == nil {
			hasTokenData = true
			// Load token report for each initiative
			for _, init := range initiatives {
				tr, trErr := report.GenerateInitiativeTokenReport(ctx, svc.Store, tokenSource, init.ID)
				if trErr == nil && tr.Totals.TotalTokens > 0 {
					tokenReports[init.ID] = tr
				}
			}
			// Calculate totals
			if len(tokenReports) > 0 {
				totalTokens = &report.TokenTotals{}
				for _, tr := range tokenReports {
					totalTokens.InputTokens += tr.Totals.InputTokens
					totalTokens.OutputTokens += tr.Totals.OutputTokens
					totalTokens.CacheReadTokens += tr.Totals.CacheReadTokens
					totalTokens.CacheCreationTokens += tr.Totals.CacheCreationTokens
					totalTokens.TotalTokens += tr.Totals.TotalTokens
					totalTokens.CostUSD += tr.Totals.CostUSD
				}
			}
		} else {
			tokenDataNote = "No omnidevx events directory found" //nolint:gosec // G101 false positive: not a credential
		}
	} else {
		tokenDataNote = "Token data unavailable" //nolint:gosec // G101 false positive: not a credential
	}

	globalStatusCounts := map[string]int{}
	var allInits []initData

	for _, init := range initiatives {
		phases, err := svc.Store.ListPhases(ctx, init.ID)
		if err != nil {
			return nil, fmt.Errorf("list phases for %s: %w", init.ID, err)
		}
		sort.Slice(phases, func(i, j int) bool {
			return phases[i].SequenceNumber < phases[j].SequenceNumber
		})

		rmis, err := svc.Store.ListRMIs(ctx, init.ID)
		if err != nil {
			return nil, fmt.Errorf("list rmis for %s: %w", init.ID, err)
		}

		// Build RMI token lookup from token report
		rmiTokens := make(map[string]*report.TokenTotals)
		phaseTokens := make(map[string]*report.TokenTotals)
		var initTokens *report.TokenTotals
		var initTokenReport *report.TokenReport
		if tr, ok := tokenReports[init.ID]; ok {
			initTokens = &tr.Totals
			initTokenReport = tr
			for _, rmiT := range tr.ByRMI {
				t := rmiT.Totals // copy
				rmiTokens[rmiT.RMIID] = &t
				// Aggregate to phase
				if phaseTokens[rmiT.PhaseID] == nil {
					phaseTokens[rmiT.PhaseID] = &report.TokenTotals{}
				}
				phaseTokens[rmiT.PhaseID].InputTokens += t.InputTokens
				phaseTokens[rmiT.PhaseID].OutputTokens += t.OutputTokens
				phaseTokens[rmiT.PhaseID].CacheReadTokens += t.CacheReadTokens
				phaseTokens[rmiT.PhaseID].CacheCreationTokens += t.CacheCreationTokens
				phaseTokens[rmiT.PhaseID].TotalTokens += t.TotalTokens
				phaseTokens[rmiT.PhaseID].CostUSD += t.CostUSD
			}
		}

		repoCounts := map[string]int{}
		rmisByPhase := map[string][]dashboardRMI{}
		totalRMIs := 0
		completedRMIs := 0
		for _, r := range rmis {
			totalRMIs++
			globalStatusCounts[strings.ToLower(r.Status)]++
			if strings.ToLower(r.Status) == "completed" {
				completedRMIs++
			}
			repoCounts[r.RepositoryID]++
			rd := dashboardRMI{RMI: r, Tokens: rmiTokens[r.ID]}
			if a, ok := assignmentsByRMI[r.ID]; ok {
				rd.ClaimedAt = a.CreatedAt.Format("2006-01-02 15:04")
				rd.ClaimedBy = a.Worker
				rd.Tooltip = "Claimed: " + rd.ClaimedAt
				if a.Worker != "" {
					rd.Tooltip += " by " + a.Worker
				}
				if r.CompletedAt != nil {
					rd.Tooltip += "\nCompleted: " + r.CompletedAt.Format("2006-01-02 15:04")
				}
			} else {
				rd.Tooltip = "Created: " + r.CreatedAt.Format("2006-01-02 15:04")
				if r.CompletedAt != nil {
					rd.Tooltip += "\nCompleted: " + r.CompletedAt.Format("2006-01-02 15:04")
				}
			}
			rmisByPhase[r.PhaseID] = append(rmisByPhase[r.PhaseID], rd)
		}
		for _, items := range rmisByPhase {
			sort.Slice(items, func(i, j int) bool {
				return items[i].RMI.SequenceNumber < items[j].RMI.SequenceNumber
			})
		}

		var repos []repoCount
		for repo, count := range repoCounts {
			repos = append(repos, repoCount{Name: shortRepo(repo), Count: count})
		}
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Count > repos[j].Count
		})

		var pds []phaseData
		for _, p := range phases {
			pds = append(pds, phaseData{Phase: p, RMIs: rmisByPhase[p.ID], Tokens: phaseTokens[p.ID]})
		}
		allInits = append(allInits, initData{
			Initiative:    init,
			Phases:        pds,
			Repos:         repos,
			TotalRMIs:     totalRMIs,
			CompletedRMIs: completedRMIs,
			Tokens:        initTokens,
			TokenReport:   initTokenReport,
		})
	}

	allDeps, err := svc.Store.ListAllDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}

	initDeps, err := svc.Store.ListAllInitiativeDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiative dependencies: %w", err)
	}

	progs, err := svc.Store.ListPrograms(ctx)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	programByID := map[string]*store.Program{}
	for _, p := range progs {
		programByID[p.ID] = p
	}

	for i := range allInits {
		if pid := allInits[i].Initiative.ProgramID; pid != "" {
			if p, ok := programByID[pid]; ok {
				allInits[i].ProgramName = p.Name
			} else {
				allInits[i].ProgramName = pid
			}
		}
	}

	progInitMap := map[string][]initData{}
	var standalone []initData
	for _, id := range allInits {
		if id.Initiative.ProgramID != "" {
			progInitMap[id.Initiative.ProgramID] = append(progInitMap[id.Initiative.ProgramID], id)
		} else {
			standalone = append(standalone, id)
		}
	}

	var programs []programData
	for pid, inits := range progInitMap {
		name := pid
		description := ""
		hidden := false
		if p, ok := programByID[pid]; ok {
			name = p.Name
			description = p.Description
			hidden = p.Hidden
		}
		// Aggregate tokens for program
		var progTokens *report.TokenTotals
		for _, ini := range inits {
			if ini.Tokens != nil {
				if progTokens == nil {
					progTokens = &report.TokenTotals{}
				}
				progTokens.InputTokens += ini.Tokens.InputTokens
				progTokens.OutputTokens += ini.Tokens.OutputTokens
				progTokens.CacheReadTokens += ini.Tokens.CacheReadTokens
				progTokens.CacheCreationTokens += ini.Tokens.CacheCreationTokens
				progTokens.TotalTokens += ini.Tokens.TotalTokens
				progTokens.CostUSD += ini.Tokens.CostUSD
			}
		}
		programs = append(programs, programData{ID: pid, Name: name, Description: description, Hidden: hidden, Initiatives: inits, Tokens: progTokens})
	}
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Name < programs[j].Name
	})

	statusOrder := map[string]int{
		"proposed": 0, "ready": 1, "in_progress": 2, "completed": 3, "cancelled": 4,
	}
	var statusDist []statusCount
	for status, count := range globalStatusCounts {
		statusDist = append(statusDist, statusCount{Status: status, Count: count})
	}
	sort.Slice(statusDist, func(i, j int) bool {
		return statusOrder[statusDist[i].Status] < statusOrder[statusDist[j].Status]
	})

	return &dashboardData{
		Initiatives:   allInits,
		Programs:      programs,
		Standalone:    standalone,
		AllDeps:       allDeps,
		InitDeps:      initDeps,
		StatusDist:    statusDist,
		TotalTokens:   totalTokens,
		HasTokenData:  hasTokenData,
		TokenDataNote: tokenDataNote,
	}, nil
}

// --- template helpers ---

var tmplFuncs = template.FuncMap{
	"statusColor":       statusColor,
	"displayStatus":     func(s string) string { return strings.ReplaceAll(s, "_", " ") },
	"typeIcon":          typeIcon,
	"shortRepo":         shortRepo,
	"safeHTML":          func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec // trusted server-side rendered HTML
	"add":               func(a, b int) int { return a + b },
	"phaseStatusCounts": phaseStatusCounts,
	"phaseTooltip":      phaseTooltip,
	"pct": func(num, denom int) int {
		if denom == 0 {
			return 0
		}
		return num * 100 / denom
	},
	"urlEncode":    url.PathEscape,
	"formatTokens": formatTokens,
	"formatCost":   formatCost,
}

func statusColor(status string) string {
	switch strings.ToLower(status) {
	case "completed", "delivery_complete", "released", "closed":
		return "#22c55e"
	case "executing", "in_progress":
		return "#3b82f6"
	case "ready":
		return "#f59e0b"
	case "planned":
		return "#8b5cf6"
	case "proposed":
		return "#6b7280"
	case "cancelled":
		return "#ef4444"
	default:
		return "#6b7280"
	}
}

func typeIcon(itemType string) string {
	switch strings.ToLower(itemType) {
	case "capability":
		return "&#9733;"
	case "refactor":
		return "&#8634;"
	case "quality":
		return "&#10003;"
	case "fix":
		return "&#9888;"
	case "chore":
		return "&#9881;"
	case "spike":
		return "&#9889;"
	default:
		return "&#8226;"
	}
}

func phaseStatusCounts(rmis []dashboardRMI) []statusCount {
	order := map[string]int{
		"proposed": 0, "ready": 1, "in_progress": 2, "completed": 3, "cancelled": 4,
	}
	counts := map[string]int{}
	for i := range rmis {
		counts[strings.ToLower(rmis[i].RMI.Status)]++
	}
	var result []statusCount
	for status, count := range counts {
		result = append(result, statusCount{Status: status, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return order[result[i].Status] < order[result[j].Status]
	})
	return result
}

func phaseTooltip(rmis []dashboardRMI) string {
	if len(rmis) == 0 {
		return ""
	}
	var earliest, latestComplete string
	completedCount := 0
	for _, rd := range rmis {
		created := rd.RMI.CreatedAt.Format("2006-01-02 15:04")
		if earliest == "" || created < earliest {
			earliest = created
		}
		if rd.RMI.CompletedAt != nil {
			completedCount++
			ct := rd.RMI.CompletedAt.Format("2006-01-02 15:04")
			if ct > latestComplete {
				latestComplete = ct
			}
		}
	}
	tip := fmt.Sprintf("%d/%d completed", completedCount, len(rmis))
	if earliest != "" {
		tip += "\nFirst created: " + earliest
	}
	if latestComplete != "" {
		tip += "\nLast completed: " + latestComplete
	}
	return tip
}

func shortRepo(repoID string) string {
	parts := strings.Split(repoID, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return repoID
}

func formatTokens(n int64) string {
	if n == 0 {
		return "-"
	}
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatCost(usd float64) string {
	if usd == 0 {
		return "-"
	}
	if usd >= 1000 {
		return fmt.Sprintf("$%.1fK", usd/1000)
	}
	if usd >= 1 {
		return fmt.Sprintf("$%.2f", usd)
	}
	return fmt.Sprintf("$%.4f", usd)
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "linux":
		cmd = exec.Command("xdg-open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		return fmt.Errorf("unsupported platform %s — open %s manually", runtime.GOOS, u)
	}
	return cmd.Start()
}

// --- renderers ---

func renderSummary(data *dashboardData) ([]byte, error) {
	tmpl := template.Must(template.New("summary").Funcs(tmplFuncs).Parse(summaryHTML))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render summary: %w", err)
	}
	return []byte(buf.String()), nil
}

func renderInitiativeDetail(data *dashboardData, id string) ([]byte, error) {
	var target *initData
	for i := range data.Initiatives {
		if data.Initiatives[i].Initiative.ID == id {
			target = &data.Initiatives[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("initiative %s not found", id)
	}

	var relevantDeps []*store.RMIDependency
	rmiSet := map[string]bool{}
	for _, p := range target.Phases {
		for _, r := range p.RMIs {
			rmiSet[r.RMI.ID] = true
		}
	}
	for _, d := range data.AllDeps {
		if rmiSet[d.SourceRMIID] || rmiSet[d.TargetRMIID] {
			relevantDeps = append(relevantDeps, d)
		}
	}

	var relevantInitDeps []*store.InitiativeDependency
	for _, d := range data.InitDeps {
		if d.SourceInitiativeID == id || d.TargetInitiativeID == id {
			relevantInitDeps = append(relevantInitDeps, d)
		}
	}

	viewData := struct {
		Init     *initData
		RMIDeps  []*store.RMIDependency
		InitDeps []*store.InitiativeDependency
	}{
		Init:     target,
		RMIDeps:  relevantDeps,
		InitDeps: relevantInitDeps,
	}

	tmpl := template.Must(template.New("detail").Funcs(tmplFuncs).Parse(detailHTML))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, viewData); err != nil {
		return nil, fmt.Errorf("render detail: %w", err)
	}
	return []byte(buf.String()), nil
}

func renderProgramDetail(data *dashboardData, id string) ([]byte, error) {
	var target *programData
	for i := range data.Programs {
		if data.Programs[i].ID == id {
			target = &data.Programs[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("program %q not found", id)
	}

	initIDs := map[string]bool{}
	for _, id := range target.Initiatives {
		initIDs[id.Initiative.ID] = true
	}

	var programInitDeps []*store.InitiativeDependency
	for _, d := range data.InitDeps {
		if initIDs[d.SourceInitiativeID] || initIDs[d.TargetInitiativeID] {
			programInitDeps = append(programInitDeps, d)
		}
	}

	viewData := struct {
		Program  *programData
		InitDeps []*store.InitiativeDependency
	}{
		Program:  target,
		InitDeps: programInitDeps,
	}

	tmpl := template.Must(template.New("program").Funcs(tmplFuncs).Parse(programHTML))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, viewData); err != nil {
		return nil, fmt.Errorf("render program: %w", err)
	}
	return []byte(buf.String()), nil
}
