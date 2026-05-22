package main

/*
vulkanSpecIRRawValue is a registry enum entry before alias targets are resolved to literal values.
*/
type vulkanSpecIRRawValue struct {
	Key         string
	Value       string
	AliasTarget string
	Doc         string
}

func vulkanSpecIRRawValuesResolveEnumValues(entries []vulkanSpecIRRawValue) []VulkanSpecIREnumValue {
	return vulkanSpecIRRawValuesResolve(entries, func(entry vulkanSpecIRRawValue) VulkanSpecIREnumValue {
		return VulkanSpecIREnumValue{
			Key:    entry.Key,
			Value:  entry.Value,
			XMLDoc: entry.Doc,
		}
	})
}

func vulkanSpecIRRawValuesResolveFlagsValues(entries []vulkanSpecIRRawValue) []VulkanSpecIRFlagsValue {
	return vulkanSpecIRRawValuesResolve(entries, func(entry vulkanSpecIRRawValue) VulkanSpecIRFlagsValue {
		return VulkanSpecIRFlagsValue{
			Key:    entry.Key,
			Value:  entry.Value,
			XMLDoc: entry.Doc,
		}
	})
}

func vulkanSpecIRRawValuesResolve[T any](entries []vulkanSpecIRRawValue, build func(vulkanSpecIRRawValue) T) []T {
	if len(entries) == 0 {
		return nil
	}

	valuesByName := make(map[string]string, len(entries))
	docsByName := make(map[string]string, len(entries))
	aliases := make([]vulkanSpecIRRawValue, 0)
	result := make([]T, 0, len(entries))

	for _, entry := range entries {
		if entry.Key == "" {
			continue
		}
		if entry.AliasTarget != "" {
			aliases = append(aliases, entry)
			continue
		}
		if entry.Value == "" {
			continue
		}
		valuesByName[entry.Key] = entry.Value
		docsByName[entry.Key] = entry.Doc
		result = append(result, build(entry))
	}

	for _, entry := range aliases {
		value, ok := valuesByName[entry.AliasTarget]
		if !ok {
			continue
		}
		doc := entry.Doc
		if doc == "" {
			doc = docsByName[entry.AliasTarget]
		}
		result = append(result, build(vulkanSpecIRRawValue{
			Key:   entry.Key,
			Value: value,
			Doc:   doc,
		}))
	}

	return result
}
