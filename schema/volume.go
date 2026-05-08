package schema

import (
	"text/template"
)

// GenerateVolumeOptions generates all volume schema options
func GetVolumeOptions() []SchemaOption {
	options := []SchemaOption{
		optVolumeName(),
		optDriver(),
		optCopy(),
		optDevice(),
		optType(),
		optImage(),
		optOptions(),
		optVolumeUser(),
		optVolumeGroup(),
		optVolumeLabel(),
		optPodmanArgsVolume(),
		optGlobalArgsVolume(),
		optContainersConfModuleVolume(),
	}

	// Pre-parse templates for all options to catch errors early. Will panic if any template is invalid, which is desirable during development.
	for _, option := range options {
		option.QuadletTemplateParsed = template.Must(template.New("quadlet").Parse(option.QuadletTemplate))
		option.PodmanTemplateParsed = template.Must(template.New("podman").Parse(option.PodmanTemplate))
	}

	return options
}

// Helper functions for volume options
func optVolumeName() SchemaOption {
	return SchemaOption{
		QuadletKey:      "VolumeName",
		PodmanKey:       "volume",
		Description:     "The (optional) name of the Podman volume. Default is systemd-<unitname> to avoid conflicts.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "{{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optCopy() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Copy",
		PodmanKey:       "--opt copy",
		Description:     "If enabled, copy content from image at mount point to the volume on first run. Default: true.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt copy",
		AllowMultiple:   false,
		Values: []OptionValue{
			{Value: "true", Description: "Copy image content to volume"},
			{Value: "false", Description: "Do not copy content"},
		},
	}
}

func optDevice() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Device",
		PodmanKey:       "--opt device=",
		Description:     "The path of a device to be mounted for the volume.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt device={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optType() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Type",
		PodmanKey:       "--opt type=",
		Description:     "The filesystem type of Device (used with mount command -t option).",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt type={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optImage() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Image",
		PodmanKey:       "--opt image=",
		Description:     "The image the volume is based on when Driver=image. Use fully qualified image names for performance. Supports :tag or digests.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt image={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optVolumeUser() SchemaOption {
	return SchemaOption{
		QuadletKey:      "User",
		PodmanKey:       "--opt uid=",
		Description:     "The host (numeric) UID or user name to use as the owner for the volume.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt uid={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optVolumeGroup() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Group",
		PodmanKey:       "--opt gid=",
		Description:     "The host (numeric) GID or group name to use as the group for the volume.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt gid={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optVolumeLabel() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Label",
		PodmanKey:       "--label",
		Description:     "Set one or more OCI labels on the volume. Format: key=value. Similar to Environment. Can be specified multiple times.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "{{.Key}} {{.Value}}",
		AllowMultiple:   true,
		Values:          []OptionValue{},
	}
}

func optOptions() SchemaOption {
	return SchemaOption{
		QuadletKey:      "Options",
		PodmanKey:       "--opt",
		Description:     "Mount options to use for filesystem (used by mount command -o option).",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "--opt o={{.Value}}",
		AllowMultiple:   false,
		Values:          []OptionValue{},
	}
}

func optPodmanArgsVolume() SchemaOption {
	return SchemaOption{
		QuadletKey:      "PodmanArgs",
		PodmanKey:       "podman",
		Description:     "Arguments passed to end of 'podman volume create' command.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "{{.Value}}",
		AllowMultiple:   true,
		Values:          []OptionValue{},
	}
}

func optGlobalArgsVolume() SchemaOption {
	return SchemaOption{
		QuadletKey:      "GlobalArgs",
		PodmanKey:       "podman",
		Description:     "Arguments passed directly after 'podman' command.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "{{.Value}}",
		AllowMultiple:   true,
		Values:          []OptionValue{},
	}
}

func optContainersConfModuleVolume() SchemaOption {
	return SchemaOption{
		QuadletKey:      "ContainersConfModule",
		PodmanKey:       "--module",
		Description:     "Load a containers.conf module. Can be specified multiple times.",
		QuadletTemplate: "{{.Key}}={{.Value}}",
		PodmanTemplate:  "{{.Key}} {{.Value}}",
		AllowMultiple:   true,
		Values:          []OptionValue{},
	}
}

// getVolumeSchema generates the complete volume schema
func GetVolumeSchema() Schema {
	options := GetVolumeOptions()
	PopulateValidators(options)

	return Schema{
		{
			Type:    "Volume",
			Options: options,
		},
	}
}
