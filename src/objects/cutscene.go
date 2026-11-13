package objects

func cutsceneIDFromProperties(properties map[string]any) string {
	raw, ok := properties["cutsceneId"]
	if !ok || raw == nil {
		return ""
	}
	id, ok := raw.(string)
	if !ok {
		return ""
	}
	return id
}
