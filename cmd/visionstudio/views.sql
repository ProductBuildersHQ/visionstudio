-- VisionStudio read-only SQL views for PRISM Control.
--
-- These views expose initiative, phase, RMI, and assignment data using
-- only base tables (no Dolt system tables) per the TRD portability rule
-- (NFR4). Column names use the Ent-generated schema; FK columns follow
-- Ent's edge-naming convention (e.g. initiative_phases, not initiative_id).
--
-- Usage:
--   prismctl db create-views          -- runs these statements against the database
--   SELECT * FROM v_initiative_summary;

-- ---------------------------------------------------------------------------
-- v_initiative_summary
-- Initiative-level rollup: status, total/completed/required-completed RMIs,
-- phase count, and distinct repos involved.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_initiative_summary AS
SELECT
    i.initiative_id,
    i.title,
    i.status,
    COUNT(DISTINCT r.rmi_id)                                          AS total_rmis,
    COUNT(DISTINCT CASE WHEN r.status = 'completed' THEN r.rmi_id END) AS completed_rmis,
    COUNT(DISTINCT CASE WHEN r.required = 1 AND r.status = 'completed'
                        THEN r.rmi_id END)                            AS required_completed,
    COUNT(DISTINCT p.phase_id)                                        AS phase_count,
    COUNT(DISTINCT r.repository_roadmap_items)                        AS repos_involved
FROM initiatives i
LEFT JOIN phases p        ON p.initiative_phases = i.initiative_id
LEFT JOIN roadmap_items r ON r.initiative_roadmap_items = i.initiative_id
GROUP BY i.initiative_id, i.title, i.status;

-- ---------------------------------------------------------------------------
-- v_phase_progress
-- Per-phase progress: total RMIs, completed count, and derived status
-- following the TRD status vocabulary:
--   completed  = all required RMIs completed
--   in_progress = any RMI in_progress
--   blocked     = any required RMI blocked
--   planned     = none started
--   partial     = all required done, optional remain
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_phase_progress AS
SELECT
    p.phase_id,
    p.initiative_phases                                                AS initiative_id,
    p.title,
    p.theme,
    p.sequence_number,
    COUNT(r.rmi_id)                                                    AS total_rmis,
    SUM(CASE WHEN r.status = 'completed' THEN 1 ELSE 0 END)           AS completed_count,
    CASE
        WHEN COUNT(r.rmi_id) = 0
            THEN 'planned'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
             AND SUM(CASE WHEN r.required = 0 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
            THEN 'completed'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
             AND SUM(CASE WHEN r.required = 0 AND r.status != 'completed' THEN 1 ELSE 0 END) > 0
            THEN 'partial'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status = 'blocked' THEN 1 ELSE 0 END) > 0
            THEN 'blocked'
        WHEN SUM(CASE WHEN r.status = 'in_progress' THEN 1 ELSE 0 END) > 0
            THEN 'in_progress'
        ELSE 'planned'
    END                                                                AS derived_status
FROM phases p
LEFT JOIN roadmap_items r ON r.phase_roadmap_items = p.phase_id
GROUP BY p.phase_id, p.initiative_phases, p.title, p.theme, p.sequence_number;

-- ---------------------------------------------------------------------------
-- v_rmi_detail
-- Flat RMI detail with denormalized initiative, phase, and repo info.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_rmi_detail AS
SELECT
    r.rmi_id,
    r.title,
    r.status,
    r.item_type,
    r.required,
    r.repository_roadmap_items                                        AS repository_id,
    r.initiative_roadmap_items                                        AS initiative_id,
    r.phase_roadmap_items                                             AS phase_id,
    r.created_at,
    r.completed_at
FROM roadmap_items r;

-- ---------------------------------------------------------------------------
-- v_active_assignments
-- Currently active work assignments with the related RMI title.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_active_assignments AS
SELECT
    a.assignment_id,
    a.roadmap_item_assignments                                        AS rmi_id,
    r.title                                                           AS rmi_title,
    a.worker,
    a.lease_expires_at,
    a.workspace
FROM assignments a
JOIN roadmap_items r ON r.rmi_id = a.roadmap_item_assignments
WHERE a.status = 'active';

-- ---------------------------------------------------------------------------
-- Token Attribution Views (TRD §16)
-- Cross-database views joining devx.token_events with prismcontrol tables.
-- These require the devx database to be populated via 'prismctl db ingest-tokens'.
-- ---------------------------------------------------------------------------

-- v_initiative_tokens
-- Token spend attributed to initiatives via assignment session matching.
-- Groups by initiative with totals and cost.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_initiative_tokens AS
SELECT
    i.initiative_id,
    i.title                                                           AS initiative_title,
    COUNT(DISTINCT t.event_id)                                        AS event_count,
    COALESCE(SUM(t.input_tokens), 0)                                  AS input_tokens,
    COALESCE(SUM(t.output_tokens), 0)                                 AS output_tokens,
    COALESCE(SUM(t.cache_read_tokens), 0)                             AS cache_read_tokens,
    COALESCE(SUM(t.cache_creation_tokens), 0)                         AS cache_creation_tokens,
    COALESCE(SUM(t.input_tokens + t.output_tokens +
                 t.cache_read_tokens + t.cache_creation_tokens), 0)   AS total_tokens
FROM prismcontrol.initiatives i
LEFT JOIN prismcontrol.roadmap_items r
    ON r.initiative_roadmap_items = i.initiative_id
LEFT JOIN prismcontrol.assignments a
    ON a.roadmap_item_assignments = r.rmi_id
LEFT JOIN devx.token_events t
    ON t.session_id = a.worker
    AND t.timestamp >= a.created_at
    AND (a.completed_at IS NULL OR t.timestamp <= a.completed_at)
GROUP BY i.initiative_id, i.title;

-- v_rmi_tokens
-- Token spend per RMI via assignment session matching.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_rmi_tokens AS
SELECT
    r.rmi_id,
    r.title                                                           AS rmi_title,
    r.initiative_roadmap_items                                        AS initiative_id,
    r.phase_roadmap_items                                             AS phase_id,
    COUNT(DISTINCT t.event_id)                                        AS event_count,
    COALESCE(SUM(t.input_tokens), 0)                                  AS input_tokens,
    COALESCE(SUM(t.output_tokens), 0)                                 AS output_tokens,
    COALESCE(SUM(t.cache_read_tokens), 0)                             AS cache_read_tokens,
    COALESCE(SUM(t.cache_creation_tokens), 0)                         AS cache_creation_tokens,
    COALESCE(SUM(t.input_tokens + t.output_tokens +
                 t.cache_read_tokens + t.cache_creation_tokens), 0)   AS total_tokens
FROM prismcontrol.roadmap_items r
LEFT JOIN prismcontrol.assignments a
    ON a.roadmap_item_assignments = r.rmi_id
LEFT JOIN devx.token_events t
    ON t.session_id = a.worker
    AND t.timestamp >= a.created_at
    AND (a.completed_at IS NULL OR t.timestamp <= a.completed_at)
GROUP BY r.rmi_id, r.title, r.initiative_roadmap_items, r.phase_roadmap_items;

-- v_unattributed_tokens
-- Token events that don't match any assignment (repository-level or unmanaged).
-- Useful for identifying coverage gaps.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_unattributed_tokens AS
SELECT
    t.event_id,
    t.timestamp,
    t.session_id,
    t.workspace,
    t.model,
    t.input_tokens,
    t.output_tokens,
    t.cache_read_tokens,
    t.cache_creation_tokens,
    t.input_tokens + t.output_tokens +
        t.cache_read_tokens + t.cache_creation_tokens                 AS total_tokens,
    CASE
        WHEN repo.repository_id IS NOT NULL THEN 'repository'
        ELSE 'unmanaged'
    END                                                               AS attribution_bucket,
    repo.repository_id
FROM devx.token_events t
LEFT JOIN prismcontrol.assignments a
    ON t.session_id = a.worker
    AND t.timestamp >= a.created_at
    AND (a.completed_at IS NULL OR t.timestamp <= a.completed_at)
LEFT JOIN prismcontrol.repositories repo
    ON t.workspace LIKE CONCAT(repo.local_path, '%')
WHERE a.assignment_id IS NULL;
