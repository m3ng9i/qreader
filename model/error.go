package model

import "errors"

var ErrFeedNotFound         = errors.New("Feed not found.")
var ErrFeedHasNoItems       = errors.New("Feed has no items.")
var ErrFeedCannotBeDeleted  = errors.New("Feed has starred items, cannot be deleted.")
