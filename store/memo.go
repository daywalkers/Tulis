package store

go

import (
    "context"
    "errors"
    "math/rand"
    "time"

    "github.com/usememos/memos/internal/util"

    storepb "github.com/usememos/memos/proto/gen/store"
)



// Visibility is the type of a visibility.
type Visibility string

const (
	// Public is the PUBLIC visibility.
	Public Visibility = "PUBLIC"
	// Protected is the PROTECTED visibility.
	Protected Visibility = "PROTECTED"
	// Private is the PRIVATE visibility.
	Private Visibility = "PRIVATE"
)

func (v Visibility) String() string {
	switch v {
	case Public:
		return "PUBLIC"
	case Protected:
		return "PROTECTED"
	case Private:
		return "PRIVATE"
	}
	return "PRIVATE"
}

go

type Memo struct {
    // ID is the system generated unique identifier for the memo.
    ID int32
    // UID is the user defined unique identifier for the memo.
    UID string
    // ShortID is the short identifier for sharing the memo (e.g., "abc1234").
    ShortID string

    // Standard fields
    RowStatus RowStatus
    CreatorID int32
    CreatedTs int64
    UpdatedTs int64

    // Domain specific fields
    Content    string
    Visibility Visibility
    Pinned     bool
    Payload    *storepb.MemoPayload

    // Composed fields
    ParentID *int32
}



type FindMemo struct {
    ID  *int32
    UID *string
    ShortID *string

    // Standard fields
    RowStatus       *RowStatus
    CreatorID       *int32
    CreatedTsAfter  *int64
    CreatedTsBefore *int64
    UpdatedTsAfter  *int64
    UpdatedTsBefore *int64

    // Domain specific fields
    ContentSearch   []string
    VisibilityList  []Visibility
    PayloadFind     *FindMemoPayload
    ExcludeContent  bool
    ExcludeComments bool
    Filter          *string

    // Pagination
    Limit  *int
    Offset *int

    // Ordering
    OrderByUpdatedTs bool
    OrderByPinned    bool
    OrderByTimeAsc   bool
}

type FindMemoPayload struct {
	Raw                *string
	TagSearch          []string
	HasLink            bool
	HasTaskList        bool
	HasCode            bool
	HasIncompleteTasks bool
}

type UpdateMemo struct {
	ID         int32
	UID        *string
	CreatedTs  *int64
	UpdatedTs  *int64
	RowStatus  *RowStatus
	Content    *string
	Visibility *Visibility
	Pinned     *bool
	Payload    *storepb.MemoPayload
}

type DeleteMemo struct {
	ID int32
}


func (s *Store) CreateMemo(ctx context.Context, create *Memo) (*Memo, error) {
    // Validate UID as before
    if !util.UIDMatcher.MatchString(create.UID) {
        return nil, errors.New("invalid uid")
    }

    // Generate a unique ShortID
    const letters = "abcdefghijklmnopqrstuvwxyz"
    const digits = "0123456789"
    rand.Seed(time.Now().UnixNano())
    for {
        shortID := ""
        // Generate 3 random letters
        for i := 0; i < 3; i++ {
            shortID += string(letters[rand.Intn(len(letters))])
        }
        // Generate 4 random digits
        for i := 0; i < 4; i++ {
            shortID += string(digits[rand.Intn(len(digits))])
        }

        // Check if ShortID is unique
        existing, err := s.GetMemo(ctx, &FindMemo{ShortID: &shortID})
        if err != nil {
            return nil, err
        }
        if existing == nil {
            create.ShortID = shortID
            break
        }
        // If not unique, loop will try again
    }

    // Create the memo with the ShortID
    return s.driver.CreateMemo(ctx, create)
}

func (s *Store) ListMemos(ctx context.Context, find *FindMemo) ([]*Memo, error) {
	return s.driver.ListMemos(ctx, find)
}

func (s *Store) GetMemo(ctx context.Context, find *FindMemo) (*Memo, error) {
	list, err := s.ListMemos(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	memo := list[0]
	return memo, nil
}

func (s *Store) UpdateMemo(ctx context.Context, update *UpdateMemo) error {
	if update.UID != nil && !util.UIDMatcher.MatchString(*update.UID) {
		return errors.New("invalid uid")
	}
	return s.driver.UpdateMemo(ctx, update)
}

func (s *Store) DeleteMemo(ctx context.Context, delete *DeleteMemo) error {
	return s.driver.DeleteMemo(ctx, delete)
}
