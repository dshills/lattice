package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxTitleLen       = 500
	MaxDescriptionLen = 10000
	MaxTypeLen        = 100
	MaxTagLen         = 100
	MaxTagCount       = 50
	MaxHierarchyDepth = 100
)

// RelationshipType represents a directed link type between WorkItems.
type RelationshipType string

const (
	Blocks      RelationshipType = "blocks"
	DependsOn   RelationshipType = "depends_on"
	RelatesTo   RelationshipType = "relates_to"
	DuplicateOf RelationshipType = "duplicate_of"
)

var validRelationshipTypes = map[RelationshipType]bool{
	Blocks:      true,
	DependsOn:   true,
	RelatesTo:   true,
	DuplicateOf: true,
}

// ValidRelationshipType returns true if the given type is a supported relationship type.
func ValidRelationshipType(t RelationshipType) bool {
	return validRelationshipTypes[t]
}

// Relationship is a directed link from one WorkItem to another.
type Relationship struct {
	ID       string           `json:"id"`
	Type     RelationshipType `json:"type"`
	TargetID string           `json:"target_id"`
}

// WorkItem is the fundamental unit of work in Lattice.
type WorkItem struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	State         State          `json:"state"`
	Tags          []string       `json:"tags"`
	Type          string         `json:"type,omitempty"`
	ParentID      *string        `json:"parent_id"`
	Relationships []Relationship `json:"relationships"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Validate checks field constraints for a WorkItem.
func (w *WorkItem) Validate() error {
	if strings.TrimSpace(w.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(w.Title) > MaxTitleLen {
		return fmt.Errorf("%w: title exceeds %d characters", ErrInvalidInput, MaxTitleLen)
	}
	if utf8.RuneCountInString(w.Description) > MaxDescriptionLen {
		return fmt.Errorf("%w: description exceeds %d characters", ErrInvalidInput, MaxDescriptionLen)
	}
	if utf8.RuneCountInString(w.Type) > MaxTypeLen {
		return fmt.Errorf("%w: type exceeds %d characters", ErrInvalidInput, MaxTypeLen)
	}
	if !ValidState(w.State) {
		return fmt.Errorf("%w: invalid state %q", ErrInvalidInput, w.State)
	}
	if len(w.Tags) > MaxTagCount {
		return fmt.Errorf("%w: exceeds maximum of %d tags", ErrInvalidInput, MaxTagCount)
	}
	for _, tag := range w.Tags {
		if utf8.RuneCountInString(tag) > MaxTagLen {
			return fmt.Errorf("%w: tag exceeds %d characters", ErrInvalidInput, MaxTagLen)
		}
		if strings.Contains(tag, ",") {
			return fmt.Errorf("%w: tag must not contain commas", ErrInvalidInput)
		}
	}
	if w.ParentID != nil && *w.ParentID == w.ID {
		return fmt.Errorf("%w: parent_id must not reference own ID", ErrValidation)
	}
	for _, rel := range w.Relationships {
		if !ValidRelationshipType(rel.Type) {
			return fmt.Errorf("%w: invalid relationship type %q", ErrInvalidInput, rel.Type)
		}
		if rel.TargetID == "" {
			return fmt.Errorf("%w: relationship target_id is required", ErrInvalidInput)
		}
	}
	return nil
}
