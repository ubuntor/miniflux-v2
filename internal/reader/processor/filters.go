// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"
)

func isBlockedByRules(feed *model.Feed, entry *model.Entry, all_rules string) bool {
	if all_rules == "" {
		return false
	}
	rules := strings.Split(all_rules, "\n")
	for _, rule := range rules {
		parts := strings.SplitN(rule, "=", 2)

		var match bool
		switch parts[0] {
		case "EntryDate":
			datePattern := parts[1]
			match = isDateMatchingPattern(entry.Date, datePattern)
		case "EntryTitle":
			match, _ = regexp.MatchString(parts[1], entry.Title)
		case "EntryURL":
			match, _ = regexp.MatchString(parts[1], entry.URL)
		case "EntryCommentsURL":
			match, _ = regexp.MatchString(parts[1], entry.CommentsURL)
		case "EntryContent":
			match, _ = regexp.MatchString(parts[1], entry.Content)
		case "EntryAuthor":
			match, _ = regexp.MatchString(parts[1], entry.Author)
		case "EntryTag":
			containsTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
				match, _ = regexp.MatchString(parts[1], tag)
				return match
			})
			if containsTag {
				match = true
			}
		}

		if match {
			slog.Debug("Blocking entry based on rule",
				slog.String("entry_url", entry.URL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("rule", rule),
			)
			return true
		}
	}
	return false
}

func isBlockedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	return isBlockedByRules(feed, entry, user.BlockFilterEntryRules) || isBlockedByRules(feed, entry, feed.BlocklistRules)
}

func isAllowedByRules(feed *model.Feed, entry *model.Entry, all_rules string) bool {
	rules := strings.Split(all_rules, "\n")
	for _, rule := range rules {
		parts := strings.SplitN(rule, "=", 2)

		var match bool
		switch parts[0] {
		case "EntryDate":
			datePattern := parts[1]
			match = isDateMatchingPattern(entry.Date, datePattern)
		case "EntryTitle":
			match, _ = regexp.MatchString(parts[1], entry.Title)
		case "EntryURL":
			match, _ = regexp.MatchString(parts[1], entry.URL)
		case "EntryCommentsURL":
			match, _ = regexp.MatchString(parts[1], entry.CommentsURL)
		case "EntryContent":
			match, _ = regexp.MatchString(parts[1], entry.Content)
		case "EntryAuthor":
			match, _ = regexp.MatchString(parts[1], entry.Author)
		case "EntryTag":
			containsTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
				match, _ = regexp.MatchString(parts[1], tag)
				return match
			})
			if containsTag {
				match = true
			}
		}

		if match {
			slog.Debug("Allowing entry based on rule",
				slog.String("entry_url", entry.URL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("rule", rule),
			)
			return true
		}
	}
	return false
}

func isAllowedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if user.KeepFilterEntryRules != "" {
		return isAllowedByRules(feed, entry, user.KeepFilterEntryRules)
	}

	if feed.KeeplistRules != "" {
		return isAllowedByRules(feed, entry, feed.KeeplistRules)
	}

	return true
}

func isDateMatchingPattern(entryDate time.Time, pattern string) bool {
	if pattern == "future" {
		return entryDate.After(time.Now())
	}

	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) != 2 {
		return false
	}

	operator := parts[0]
	dateStr := parts[1]

	switch operator {
	case "before":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.Before(targetDate)
	case "after":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.After(targetDate)
	case "between":
		dates := strings.Split(dateStr, ",")
		if len(dates) != 2 {
			return false
		}
		startDate, err1 := time.Parse("2006-01-02", dates[0])
		endDate, err2 := time.Parse("2006-01-02", dates[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return entryDate.After(startDate) && entryDate.Before(endDate)
	}
	return false
}
