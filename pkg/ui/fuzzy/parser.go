package fuzzy

import (
	"regexp"
	"strings"
)

// ParseSearchQuery parses a raw search query into structured filters
func ParseSearchQuery(raw string) *SearchQuery {
	query := &SearchQuery{
		Raw:             strings.TrimSpace(raw),
		Terms:           []string{},
		Filters:         make(map[string]string),
		TagFilters:      make(map[string]string),
		NegativeFilters: []string{},
		IPFilters:       []string{},
		DNSFilters:      []string{},
		HasTags:         []string{},
		MissingTags:     []string{},
	}

	if query.Raw == "" {
		return query
	}

	// Split by spaces but respect quoted strings
	parts := parseQuotedStrings(query.Raw)

	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "!"):
			// Negative filter
			query.NegativeFilters = append(query.NegativeFilters, part[1:])
		case strings.Contains(part, ":"):
			// Structured filter
			key, value := splitKeyValue(part)
			switch key {
			case "name":
				query.Filters["name"] = value
			case "id":
				query.Filters["instance-id"] = value
			case "ip", "private-ip", "public-ip":
				query.IPFilters = append(query.IPFilters, value)
			case "dns", "private-dns", "public-dns":
				query.DNSFilters = append(query.DNSFilters, value)
			case "state":
				query.StateFilter = value
			case "type":
				query.TypeFilter = value
			case "az", "zone":
				query.AZFilter = value
			case "tag":
				if tagKey, tagValue := splitKeyValue(value); tagKey != "" {
					query.TagFilters[tagKey] = tagValue
				}
			case "has":
				query.HasTags = append(query.HasTags, value)
			case "missing":
				query.MissingTags = append(query.MissingTags, value)
			}
		default:
			// Fuzzy search term
			query.Terms = append(query.Terms, part)
		}
	}

	return query
}

// parseQuotedStrings splits a string by spaces while respecting quoted strings
func parseQuotedStrings(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range s {
		switch {
		case escapeNext:
			current.WriteRune(r)
			escapeNext = false
		case r == '\\':
			escapeNext = true
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// splitKeyValue splits a key:value pair, handling the case where value contains colons
func splitKeyValue(pair string) (string, string) {
	parts := strings.SplitN(pair, ":", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// MatchesQuery checks if an instance matches the search query
func (i *Instance) MatchesQuery(query *SearchQuery) bool {
	// Check negative filters first
	for _, filter := range query.NegativeFilters {
		if i.matchesNegativeFilter(filter) {
			return false
		}
	}

	// Check exact filters
	for key, value := range query.Filters {
		if !i.matchesFilter(key, value) {
			return false
		}
	}

	// Check tag filters
	for tagKey, tagValue := range query.TagFilters {
		if instanceValue, exists := i.Tags[tagKey]; !exists || instanceValue != tagValue {
			return false
		}
	}

	// Check IP filters
	for _, ipPattern := range query.IPFilters {
		if !i.matchesIPPattern(ipPattern) {
			return false
		}
	}

	// Check DNS filters
	for _, dnsPattern := range query.DNSFilters {
		if !i.matchesDNSPattern(dnsPattern) {
			return false
		}
	}

	// Check state filter
	if query.StateFilter != "" && !strings.EqualFold(i.State, query.StateFilter) {
		return false
	}

	// Check type filter
	if query.TypeFilter != "" && !matchesPattern(i.InstanceType, query.TypeFilter) {
		return false
	}

	// Check AZ filter
	if query.AZFilter != "" && !strings.EqualFold(i.AvailabilityZone, query.AZFilter) {
		return false
	}

	// Check has tags
	for _, tag := range query.HasTags {
		if _, exists := i.Tags[tag]; !exists {
			return false
		}
	}

	// Check missing tags
	for _, tag := range query.MissingTags {
		if _, exists := i.Tags[tag]; exists {
			return false
		}
	}

	// Check fuzzy terms
	if len(query.Terms) > 0 && !i.matchesFuzzyTerms(query.Terms) {
		return false
	}

	return true
}

// matchesNegativeFilter checks if the instance matches a negative filter
func (i *Instance) matchesNegativeFilter(filter string) bool {
	key, value := splitKeyValue(filter)
	switch key {
	case "name":
		return strings.Contains(strings.ToLower(i.Name), strings.ToLower(value))
	case "id":
		return strings.Contains(strings.ToLower(i.InstanceID), strings.ToLower(value))
	case "state":
		return strings.EqualFold(i.State, value)
	case "type":
		return matchesPattern(i.InstanceType, value)
	case "az", "zone":
		return strings.EqualFold(i.AvailabilityZone, value)
	case "tag":
		if tagKey, tagValue := splitKeyValue(value); tagKey != "" {
			if instanceValue, exists := i.Tags[tagKey]; exists {
				return strings.Contains(strings.ToLower(instanceValue), strings.ToLower(tagValue))
			}
		}
	}
	return false
}

// matchesFilter checks if the instance matches an exact filter
func (i *Instance) matchesFilter(key, value string) bool {
	switch key {
	case "name":
		return strings.Contains(strings.ToLower(i.Name), strings.ToLower(value))
	case "instance-id":
		return strings.Contains(strings.ToLower(i.InstanceID), strings.ToLower(value))
	}
	return false
}

// matchesIPPattern checks if the instance matches an IP pattern
func (i *Instance) matchesIPPattern(pattern string) bool {
	if matchesPattern(i.PrivateIP, pattern) {
		return true
	}
	if matchesPattern(i.PublicIP, pattern) {
		return true
	}
	return false
}

// matchesDNSPattern checks if the instance matches a DNS pattern
func (i *Instance) matchesDNSPattern(pattern string) bool {
	if matchesPattern(i.PrivateDNS, pattern) {
		return true
	}
	if matchesPattern(i.PublicDNS, pattern) {
		return true
	}
	return false
}

// matchesFuzzyTerms checks if the instance matches all of the fuzzy terms
func (i *Instance) matchesFuzzyTerms(terms []string) bool {
	if len(terms) == 0 {
		return true // No terms means everything matches
	}

	for _, term := range terms {
		termLower := strings.ToLower(term)
		matched := strings.Contains(strings.ToLower(i.Name), termLower) ||
			strings.Contains(strings.ToLower(i.InstanceID), termLower) ||
			strings.Contains(strings.ToLower(i.PrivateIP), termLower) ||
			strings.Contains(strings.ToLower(i.PublicIP), termLower)

		// Check tags
		if !matched {
			for _, tagValue := range i.Tags {
				if strings.Contains(strings.ToLower(tagValue), termLower) {
					matched = true
					break
				}
			}
		}

		if !matched {
			return false // All terms must match
		}
	}
	return true
}

// matchesPattern checks if a string matches a pattern (supports wildcards)
func matchesPattern(str, pattern string) bool {
	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*")
	regexPattern = "^" + regexPattern + "$"
	matched, err := regexp.MatchString(regexPattern, str)
	if err != nil {
		// If regex is invalid, fall back to exact match
		return str == pattern
	}
	return matched
}

// CalculateScore calculates a fuzzy matching score for the instance
func (i *Instance) CalculateScore(query *SearchQuery, weights WeightConfig) float64 {
	var score float64

	// Score based on fuzzy terms
	for _, term := range query.Terms {
		termLower := strings.ToLower(term)

		// Name matches
		if strings.Contains(strings.ToLower(i.Name), termLower) {
			score += float64(weights.Name)
		}

		// Instance ID matches
		if strings.Contains(strings.ToLower(i.InstanceID), termLower) {
			score += float64(weights.InstanceID)
		}

		// Tag matches
		for _, tagValue := range i.Tags {
			if strings.Contains(strings.ToLower(tagValue), termLower) {
				score += float64(weights.Tags)
				break
			}
		}

		// IP matches
		if strings.Contains(strings.ToLower(i.PrivateIP), termLower) ||
			strings.Contains(strings.ToLower(i.PublicIP), termLower) {
			score += float64(weights.IP)
		}

		// DNS matches
		if strings.Contains(strings.ToLower(i.PrivateDNS), termLower) ||
			strings.Contains(strings.ToLower(i.PublicDNS), termLower) {
			score += float64(weights.DNS)
		}
	}

	// Bonus for exact matches in filters
	if name, ok := query.Filters["name"]; ok && strings.EqualFold(i.Name, name) {
		score += float64(weights.Name * 2) // Double bonus for exact match
	}

	if id, ok := query.Filters["instance-id"]; ok && strings.EqualFold(i.InstanceID, id) {
		score += float64(weights.InstanceID * 2)
	}

	// Tag filter exact matches
	for tagKey, tagValue := range query.TagFilters {
		if instanceValue, exists := i.Tags[tagKey]; exists && strings.EqualFold(instanceValue, tagValue) {
			score += float64(weights.Tags * 2)
		}
	}

	return score
}
