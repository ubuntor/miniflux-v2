// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"miniflux.app/v2/internal/locale"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// ValidateRange makes sure the offset/limit values are valid.
func ValidateRange(offset, limit int) error {
	if offset < 0 {
		return fmt.Errorf(`offset value should be >= 0`)
	}

	if limit < 0 {
		return fmt.Errorf(`limit value should be >= 0`)
	}

	return nil
}

// ValidateDirection makes sure the sorting direction is valid.
func ValidateDirection(direction string) error {
	switch direction {
	case "asc", "desc":
		return nil
	}

	return fmt.Errorf(`invalid direction, valid direction values are: "asc" or "desc"`)
}

// IsValidRegex verifies if the regex can be compiled.
func IsValidRegex(expr string) bool {
	_, err := regexp.Compile(expr)
	return err == nil
}

// IsValidURL verifies if the provided value is a valid absolute URL.
func IsValidURL(absoluteURL string) bool {
	_, err := url.ParseRequestURI(absoluteURL)
	return err == nil
}

func IsValidDomain(domain string) bool {
	domain = strings.ToLower(domain)

	if len(domain) < 1 || len(domain) > 253 {
		return false
	}

	return domainRegex.MatchString(domain)
}

func IsValidDomainList(value string) bool {
	domains := strings.Split(strings.TrimSpace(value), " ")
	for _, domain := range domains {
		if !IsValidDomain(domain) {
			return false
		}
	}

	return true
}

func isValidFilterRules(filterEntryRules string, filterType string) *locale.LocalizedError {
	// Valid Format: FieldName=RegEx\nFieldName=RegEx...
	fieldNames := []string{"EntryTitle", "EntryURL", "EntryCommentsURL", "EntryContent", "EntryAuthor", "EntryTag", "EntryDate"}

	rules := strings.Split(filterEntryRules, "\n")
	for i, rule := range rules {
		// Check if rule starts with a valid fieldName
		idx := slices.IndexFunc(fieldNames, func(fieldName string) bool { return strings.HasPrefix(rule, fieldName) })
		if idx == -1 {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_fieldname_invalid", i+1, "'"+strings.Join(fieldNames, "', '")+"'")
		}
		fieldName := fieldNames[idx]
		fieldRegEx, _ := strings.CutPrefix(rule, fieldName)

		// Check if regex begins with a =
		if !strings.HasPrefix(fieldRegEx, "=") {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_separator_required", i+1)
		}
		fieldRegEx = strings.TrimPrefix(fieldRegEx, "=")

		if fieldRegEx == "" {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_regex_required", i+1)
		}

		// Check if provided pattern is a valid RegEx
		if !IsValidRegex(fieldRegEx) {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_invalid_regex", i+1)
		}
	}
	return nil
}
