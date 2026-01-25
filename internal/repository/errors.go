package repository

import "errors"

var ErrProfileNotFound = errors.New("profile not found")
var ErrFailedToUpdateProfile = errors.New("failed to update profile")
var ErrUsernameTaken = errors.New("username already taken")
