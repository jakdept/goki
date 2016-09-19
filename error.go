package main

import "fmt"

type Error struct {
	Code       int
	path       string
	value      interface{}
	innerError error
}

// Upgrades an external error making it a error
func UpgradeError(e error) Error {
	return Error{Code: ErrUpgradedError, innerError: e}
}

// the function `Error` to make my custom errors work
func (e *Error) Error() string {
	var args []interface{}

	if e.path != "" {
		args = append(args, e.path)
	}

	if e.value != nil {
		args = append(args, e.value)
	}

	if e.innerError != nil {
		args = append(args, e.innerError)
	}

	if len(args) < 1 {
		return errMsg[e.Code]
	}
	return fmt.Sprintf(errMsg[e.Code], args...)
}

// assign a unique id to each error
const (
	ErrUpgradedError = 1 << iota
	ErrBadType
	ErrBadAddressStructure
	ErrBadAddressIndex
	ErrReadConfig
	ErrParseConfig
	ErrPageRead
	ErrPageNoTitle
	ErrParseTemplates
	ErrPageRestricted
	ErrIndexError
	ErrWatcherCreate
	ErrWatcherAdd
	ErrIndexCreate
	ErrIndexClose
	ErrIndexRemove
	ErrFileRead
	ErrInvalidQuery
	ErrListField
	ErrFormatSearchResponse
	ErrResultsFormatType
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrUpgradedError:        "nothing to see here",
	ErrBadType:              "value at address [%s] is of the wrong type [%s]",
	ErrBadAddressStructure:  "got an address mapping that does not match the formatting",
	ErrBadAddressIndex:      "got address mapping that does not exist",
	ErrReadConfig:           "error reading config [%s] - %v",
	ErrParseConfig:          "parse config error %v - contents %#v",
	ErrPageRead:             "error reading from file - %v",
	ErrPageNoTitle:          "read no titles on page [%s]",
	ErrParseTemplates:       "problem parsing templates - %v",
	ErrPageRestricted:       "hit a restricted page - %s",
	ErrIndexError:           "problem with index at [%s] - %v",
	ErrWatcherCreate:        "problem creating a watcher - %v",
	ErrWatcherAdd:           "problem watching a directory %s - %v",
	ErrIndexCreate:          "problem creating index at [%s] - %v",
	ErrIndexClose:           "failed to close index - %v",
	ErrIndexRemove:          "failed to remove index at %s - %v",
	ErrFileRead:             "failed to read directory - %s - %v",
	ErrInvalidQuery:         "bad query passed - %v",
	ErrListField:            "could not list field - %v",
	ErrFormatSearchResponse: "ran into an issue formatting the results - %v",
	ErrResultsFormatType:    "returned %s was of the wrong type - %#v",
}
