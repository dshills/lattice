package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func validWorkItem() *WorkItem {
	return &WorkItem{
		ID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		ProjectID: DefaultProjectID,
		Title:     "Test Item",
		State:     NotDone,
	}
}

func TestValidate_ValidWorkItem(t *testing.T) {
	w := validWorkItem()
	assert.NoError(t, w.Validate())
}

func TestValidate_TitleRequired(t *testing.T) {
	w := validWorkItem()
	w.Title = ""
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)

	w.Title = "   "
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_TitleMaxLength(t *testing.T) {
	w := validWorkItem()
	w.Title = strings.Repeat("a", MaxTitleLen)
	assert.NoError(t, w.Validate())

	w.Title = strings.Repeat("a", MaxTitleLen+1)
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_DescriptionMaxLength(t *testing.T) {
	w := validWorkItem()
	w.Description = strings.Repeat("a", MaxDescriptionLen)
	assert.NoError(t, w.Validate())

	w.Description = strings.Repeat("a", MaxDescriptionLen+1)
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_TypeMaxLength(t *testing.T) {
	w := validWorkItem()
	w.Type = strings.Repeat("a", MaxTypeLen)
	assert.NoError(t, w.Validate())

	w.Type = strings.Repeat("a", MaxTypeLen+1)
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_InvalidState(t *testing.T) {
	w := validWorkItem()
	w.State = "invalid"
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_TagCount(t *testing.T) {
	w := validWorkItem()
	w.Tags = make([]string, MaxTagCount)
	for i := range w.Tags {
		w.Tags[i] = "tag"
	}
	assert.NoError(t, w.Validate())

	w.Tags = make([]string, MaxTagCount+1)
	for i := range w.Tags {
		w.Tags[i] = "tag"
	}
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_TagMaxLength(t *testing.T) {
	w := validWorkItem()
	w.Tags = []string{strings.Repeat("a", MaxTagLen)}
	assert.NoError(t, w.Validate())

	w.Tags = []string{strings.Repeat("a", MaxTagLen+1)}
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_TagNoCommas(t *testing.T) {
	w := validWorkItem()
	w.Tags = []string{"valid-tag"}
	assert.NoError(t, w.Validate())

	w.Tags = []string{"invalid,tag"}
	assert.ErrorIs(t, w.Validate(), ErrInvalidInput)
}

func TestValidate_ParentIDNotSelf(t *testing.T) {
	w := validWorkItem()
	id := w.ID
	w.ParentID = &id
	err := w.Validate()
	assert.True(t, errors.Is(err, ErrValidation))
}

func TestValidate_ParentIDOther(t *testing.T) {
	w := validWorkItem()
	other := "other-id"
	w.ParentID = &other
	assert.NoError(t, w.Validate())
}

func TestValidRelationshipType(t *testing.T) {
	assert.True(t, ValidRelationshipType(Blocks))
	assert.True(t, ValidRelationshipType(DependsOn))
	assert.True(t, ValidRelationshipType(RelatesTo))
	assert.True(t, ValidRelationshipType(DuplicateOf))
	assert.False(t, ValidRelationshipType("parent_of"))
	assert.False(t, ValidRelationshipType(""))
}
