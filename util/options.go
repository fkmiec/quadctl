package util

import (
	. "github.com/fkmiec/quadctl/schema"
)

func GetQuadletSchemas() map[string]map[string]SchemaOption {
	//Get the schemas for each supported type
	schemas := map[string]map[string]SchemaOption{}
	schemas["volume"] = GetQuadletOptionsMap("volume")
	schemas["network"] = GetQuadletOptionsMap("network")
	schemas["container"] = GetQuadletOptionsMap("container")
	schemas["pod"] = GetQuadletOptionsMap("pod")
	return schemas
}

func GetQuadletOptionsMap(quadletType string) map[string]SchemaOption {
	var options []SchemaOption
	switch quadletType {
	case "container":
		options = GetContainerOptions()
	case "pod":
		options = GetPodOptions()
	case "network":
		options = GetNetworkOptions()
	case "volume":
		options = GetVolumeOptions()
	default:
		return nil
	}
	if options == nil {
		return nil
	}
	optionsMap := assembleQuadletOptionsMap(options)
	return optionsMap
}

func GetPodmanOptionsMap(quadletType string) map[string]SchemaOption {
	var options []SchemaOption
	switch quadletType {
	case "container":
		options = GetContainerOptions()
	case "pod":
		options = GetPodOptions()
	case "network":
		options = GetNetworkOptions()
	case "volume":
		options = GetVolumeOptions()
	default:
		return nil
	}
	if options == nil {
		return nil
	}
	optionsMap := assemblePodmanOptionsMap(options)
	return optionsMap
}

func assembleQuadletOptionsMap(options []SchemaOption) map[string]SchemaOption {
	optionsMap := make(map[string]SchemaOption)
	for _, option := range options {
		optionsMap[option.QuadletKey] = option
	}
	return optionsMap
}

func assemblePodmanOptionsMap(options []SchemaOption) map[string]SchemaOption {
	optionsMap := make(map[string]SchemaOption)
	for _, option := range options {
		optionsMap[option.PodmanKey] = option
	}
	return optionsMap
}
