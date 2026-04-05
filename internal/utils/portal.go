package utils

import "github.com/kodehat/portkey/internal/models"

// GroupPortals groups a flat portal slice into ordered PortalGroups.
// Groups appear in the order the first portal of each group is encountered.
// Portals with an empty Group field are collected into a group with an empty name
// and placed at the end, after all named groups.
func GroupPortals(portals []models.Portal) []models.PortalGroup {
	var order []string
	groups := make(map[string]*models.PortalGroup)

	for _, p := range portals {
		key := p.Group
		if _, exists := groups[key]; !exists {
			order = append(order, key)
			groups[key] = &models.PortalGroup{Name: key}
		}
		groups[key].Portals = append(groups[key].Portals, p)
	}

	// Build result: named groups first, unnamed (empty) group last.
	result := make([]models.PortalGroup, 0, len(order))
	var unnamed *models.PortalGroup
	for _, key := range order {
		if key == "" {
			unnamed = groups[key]
		} else {
			result = append(result, *groups[key])
		}
	}
	if unnamed != nil {
		result = append(result, *unnamed)
	}
	return result
}
