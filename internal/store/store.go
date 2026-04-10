package store

import (
	"context"

	"github.com/dshills/lattice/internal/domain"
)

// UpdateParams carries the fields to update on a WorkItem. Nil pointer fields
// mean "do not change". Non-nil pointer fields set the value (including empty
// string to clear a field). Tags nil = don't change; non-nil = replace entirely.
// ParentID: nil = don't change, pointer to "" = unset parent.
type UpdateParams struct {
	Title       *string
	Description *string
	State       *domain.State
	Type        *string
	Tags        []string // nil = don't change, non-nil = replace
	ParentID    *string  // nil = don't change, &"" = unset
	Override    bool
}

// ProjectUpdateParams carries the fields to update on a Project.
// Nil pointer fields mean "do not change".
type ProjectUpdateParams struct {
	Name        *string
	Description *string
}

// ProjectWithCount is a Project with a computed work item count.
type ProjectWithCount struct {
	domain.Project
	ItemCount int `json:"item_count"`
}

// ProjectStore defines persistence operations for Projects.
type ProjectStore interface {
	Create(ctx context.Context, project *domain.Project) error
	Get(ctx context.Context, id string) (*domain.Project, error)
	Update(ctx context.Context, id string, params ProjectUpdateParams) (*domain.Project, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]ProjectWithCount, error)
}

// WorkItemStore defines persistence operations for WorkItems.
// All operations are scoped to a project via the projectID parameter.
type WorkItemStore interface {
	Create(ctx context.Context, item *domain.WorkItem) error
	Get(ctx context.Context, projectID, id string) (*domain.WorkItem, error)
	Update(ctx context.Context, projectID, id string, params UpdateParams) (*domain.WorkItem, error)
	Delete(ctx context.Context, projectID, id string) error
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	AncestorDepth(ctx context.Context, parentID string) (int, error)
	HasCycle(ctx context.Context, childID, parentID string) (bool, error)
}

// RelationshipStore defines persistence operations for Relationships.
type RelationshipStore interface {
	Add(ctx context.Context, ownerID string, rel *domain.Relationship) error
	Remove(ctx context.Context, ownerID, relID string) error
	ListByTarget(ctx context.Context, targetID string) ([]domain.Relationship, error)
}

// CycleDetector detects dependency graph cycles (depends_on + blocks edges).
// Implemented by internal/graph; distinct from WorkItemStore.HasCycle which
// checks parent-child hierarchy cycles only.
type CycleDetector interface {
	DetectCycles(ctx context.Context, workItemID string) ([][]string, error)
}

// ListFilter contains query parameters for listing WorkItems.
type ListFilter struct {
	ProjectID          string
	State              *domain.State
	Tags               []string
	Type               *string
	ParentID           *string
	RelationshipType   *domain.RelationshipType
	RelationshipTarget *string
	IsBlocked          *bool
	IsReady            *bool
	Page               int
	PageSize           int
}

// ListResult contains a page of WorkItems and pagination metadata.
type ListResult struct {
	Items    []domain.WorkItem `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// UserStore defines persistence operations for Users.
type UserStore interface {
	Create(ctx context.Context, user *domain.User, passwordHash string) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateDisplayName(ctx context.Context, id, displayName string) (*domain.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	Delete(ctx context.Context, id string) error
}

// MembershipStore defines persistence operations for ProjectMemberships.
type MembershipStore interface {
	Add(ctx context.Context, membership *domain.ProjectMembership) error
	Remove(ctx context.Context, projectID, userID string) error
	UpdateRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error
	ListByProject(ctx context.Context, projectID string) ([]domain.ProjectMembership, error)
	GetRole(ctx context.Context, projectID, userID string) (domain.ProjectRole, error)
}
