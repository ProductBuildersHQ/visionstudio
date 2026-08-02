package roadmap

import (
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// DiffKind classifies a difference.
type DiffKind string

const (
	DiffStatusMismatch DiffKind = "status_mismatch"
	DiffTitleMismatch  DiffKind = "title_mismatch"
	DiffMissingInDB    DiffKind = "missing_in_db"
	DiffMissingInFile  DiffKind = "missing_in_file"
	DiffPhaseMissing   DiffKind = "phase_missing"
)

// Difference describes one discrepancy between a ROADMAP.md and the DB.
type Difference struct {
	Kind    DiffKind
	RMIID   string
	PhaseID string
	File    string // value from the file
	DB      string // value from the DB
	Message string
}

// DiffInput holds the DB-side data for comparison.
type DiffInput struct {
	Phases []GeneratePhase
}

// Diff compares a parsed ROADMAP.md against DB state and returns differences.
func Diff(parsed *Roadmap, db *DiffInput) []Difference {
	var diffs []Difference

	dbRMIs := map[string]*store.RoadmapItem{}
	dbPhaseForRMI := map[string]string{}
	for _, gp := range db.Phases {
		for _, r := range gp.RMIs {
			dbRMIs[r.ID] = r
			dbPhaseForRMI[r.ID] = gp.Phase.ID
		}
	}

	dbPhases := map[int]*GeneratePhase{}
	for i := range db.Phases {
		dbPhases[db.Phases[i].Phase.SequenceNumber] = &db.Phases[i]
	}

	fileRMIs := map[string]bool{}

	for _, fp := range parsed.Phases {
		if _, ok := dbPhases[fp.Number]; !ok {
			diffs = append(diffs, Difference{
				Kind:    DiffPhaseMissing,
				PhaseID: fmt.Sprintf("phase-%d", fp.Number),
				File:    fp.Title,
				Message: fmt.Sprintf("Phase %d (%s) exists in file but not in DB", fp.Number, fp.Title),
			})
		}

		for _, item := range fp.Items {
			fileRMIs[item.ID] = true
			dbRMI, ok := dbRMIs[item.ID]
			if !ok {
				diffs = append(diffs, Difference{
					Kind:    DiffMissingInDB,
					RMIID:   item.ID,
					File:    item.Title,
					Message: fmt.Sprintf("%s exists in file but not in DB", item.ID),
				})
				continue
			}

			fileStatus := "open"
			if item.Completed {
				fileStatus = "completed"
			}
			dbStatus := dbRMI.Status
			dbCompleted := dbStatus == "completed"

			if item.Completed != dbCompleted {
				diffs = append(diffs, Difference{
					Kind:    DiffStatusMismatch,
					RMIID:   item.ID,
					File:    fileStatus,
					DB:      dbStatus,
					Message: fmt.Sprintf("%s: file=%s, db=%s", item.ID, fileStatus, dbStatus),
				})
			}

			if item.Title != dbRMI.Title {
				diffs = append(diffs, Difference{
					Kind:    DiffTitleMismatch,
					RMIID:   item.ID,
					File:    item.Title,
					DB:      dbRMI.Title,
					Message: fmt.Sprintf("%s title: file=%q, db=%q", item.ID, item.Title, dbRMI.Title),
				})
			}
		}
	}

	for id, r := range dbRMIs {
		if !fileRMIs[id] {
			diffs = append(diffs, Difference{
				Kind:    DiffMissingInFile,
				RMIID:   id,
				DB:      r.Title,
				Message: fmt.Sprintf("%s (%s) exists in DB but not in file", id, r.Title),
			})
		}
	}

	return diffs
}
