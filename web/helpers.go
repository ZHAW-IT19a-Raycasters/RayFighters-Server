package web

import "fmt"

// ErrParseCookie is the error that happens on a parse error
type ErrParseCookie struct {
	logMsg, httpMsg string
	status          int
}

// Error satisfies the error interface
func (err ErrParseCookie) Error() string {
	return fmt.Sprintf("[%d] - [%s]", err.status, err.logMsg)
}

// ErrParseCookie is the error that happens on a parse error
type ErrNoCookie struct {
	logMsg, httpMsg string
	status          int
}

// Error satisfies the error interface
func (err ErrNoCookie) Error() string {
	return fmt.Sprintf("[%d] - [%s]", err.status, err.logMsg)
}

// ErrAuthManagerFailure is the error that is returned once
// there is an authentication error of any kind.
type ErrAuthManagerFailure struct {
	internal error
}

// Error satisfies the error interface
func (err ErrAuthManagerFailure) Error() string {
	return fmt.Sprintf("an error happened while creating the auth manager [%s]", err.Unwrap().Error())
}

// Is satisfies the errors.Is interface
func (err ErrAuthManagerFailure) Is(target error) bool {
	_, ok := target.(ErrAuthManagerFailure)
	return ok
}

// Unwrap satisfies the errors.Unwrap functionality.
func (err ErrAuthManagerFailure) Unwrap() error {
	return err.internal
}

// ErrUsernameInvalidUtf8 is the error that is returned once
// there is an invalid rune in the given user name.
type ErrUsernameInvalidUtf8 struct {
	name string
}

// Error satisfies the error interface
func (err ErrUsernameInvalidUtf8) Error() string {
	return fmt.Sprintf("invalid utf-8 runes in username %s", err.name)
}

// Is satisfies the errors.Is interface
func (err ErrUsernameInvalidUtf8) Is(target error) bool {
	_, ok := target.(ErrUsernameInvalidUtf8)
	return ok
}

// ErrUsernameInvalidLength is the error that is returned once
// the username is of invalid length, such as 0 or longer then 255.
type ErrUsernameInvalidLength struct {
	name    string
	runeLen int
}

// Error satisfies the error interface
func (err ErrUsernameInvalidLength) Error() string {
	return fmt.Sprintf("username contains too many runes %d, %s", err.runeLen, err.name)
}

// Is satisfies the errors.Is interface
func (err ErrUsernameInvalidLength) Is(target error) bool {
	_, ok := target.(ErrUsernameInvalidLength)
	return ok
}
