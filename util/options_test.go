package util

import (
	"fmt"
	"testing"

	. "github.com/fkmiec/quadctl/schema"
)

func TestVolumeQuadletOptionsToPodmanTableDriven(t *testing.T) {
	quadletSchemas := map[string]map[string]SchemaOption{}
	quadletSchemas["volume"] = GetQuadletOptionsMap("volume")
	quadletSchemas["network"] = GetQuadletOptionsMap("network")
	quadletSchemas["container"] = GetQuadletOptionsMap("container")
	quadletSchemas["pod"] = GetQuadletOptionsMap("pod")

	// Defining the columns of the table
	var tests = []struct {
		qType   string
		options map[string]SchemaOption
		key     string
		value   string
		want    string
	}{
		// the table itself
		{"volume", quadletSchemas["volume"], "ContainersConfModule", "/etc/nvd.conf", "--module=/etc/nvd.conf"},
		{"volume", quadletSchemas["volume"], "Copy", "true", "--opt copy"},
		{"volume", quadletSchemas["volume"], "Device", "tmpfs", "--opt device=tmpfs"},
		{"volume", quadletSchemas["volume"], "Driver", "image", "--driver=image"},
		{"volume", quadletSchemas["volume"], "GlobalArgs", "--log-level=debug", "--log-level=debug"},
		{"volume", quadletSchemas["volume"], "Group", "192", "--gid 192"},
		{"volume", quadletSchemas["volume"], "Image", "quay.io/centos/centos:latest", "--opt image=quay.io/centos/centos:latest"},
		{"volume", quadletSchemas["volume"], "Label", "\"foo=bar\"", "--label \"foo=bar\""},
		{"volume", quadletSchemas["volume"], "Label", "foo=bar", "--label foo=bar"},
		{"volume", quadletSchemas["volume"], "Options", "XYZ", "--opt o=XYZ"},
		{"volume", quadletSchemas["volume"], "PodmanArgs", "--driver=image --privileged", "--driver=image --privileged"},
		{"volume", quadletSchemas["volume"], "Type", "type", "--opt type=type"},
		{"volume", quadletSchemas["volume"], "User", "123", "--uid 123"},
		//{"volume", quadletSchemas["volume"], "VolumeName", "foo", "podman volume create foo"},
	}
	// The execution loop
	for _, tt := range tests {
		quadletOpt := fmt.Sprintf("%s=%s", tt.key, tt.value)
		t.Run(quadletOpt, func(t *testing.T) {
			ans, err := QuadletOptionToPodman(tt.qType, tt.options, tt.key, tt.value)
			if err != nil {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestNetworkQuadletOptionsToPodmanTableDriven(t *testing.T) {
	quadletSchemas := map[string]map[string]SchemaOption{}
	quadletSchemas["volume"] = GetQuadletOptionsMap("volume")
	quadletSchemas["network"] = GetQuadletOptionsMap("network")
	quadletSchemas["container"] = GetQuadletOptionsMap("container")
	quadletSchemas["pod"] = GetQuadletOptionsMap("pod")

	// Defining the columns of the table
	var tests = []struct {
		qType   string
		options map[string]SchemaOption
		key     string
		value   string
		want    string
	}{
		// the table itself
		{"network", quadletSchemas["network"], "ContainersConfModule", "/etc/nvd.conf", "--module=/etc/nvd.conf"},
		{"network", quadletSchemas["network"], "DisableDNS", "true", "--disable-dns"},
		{"network", quadletSchemas["network"], "DNS", "192.168.55.1", "--dns=192.168.55.1"},
		{"network", quadletSchemas["network"], "Driver", "bridge", "--driver bridge"},
		{"network", quadletSchemas["network"], "Gateway", "192.168.55.3", "--gateway 192.168.55.3"},
		{"network", quadletSchemas["network"], "GlobalArgs", "--log-level=debug", "--log-level=debug"},
		{"network", quadletSchemas["network"], "InterfaceName", "enp1", "--interface-name enp1"},
		{"network", quadletSchemas["network"], "Internal", "true", "--internal"},
		{"network", quadletSchemas["network"], "IPAMDriver", "dhcp", "--ipam-driver dhcp"},
		{"network", quadletSchemas["network"], "IPRange", "192.168.55.128/25", "--ip-range 192.168.55.128/25"},
		{"network", quadletSchemas["network"], "IPv6", "true", "--ipv6"},
		{"network", quadletSchemas["network"], "Label", "\"XYZ\"", "--label \"XYZ\""},
		//{"network", quadletSchemas["network"], "NetworkDeleteOnStop", "true", "Add ExecStopPost to delete the network when the unit is stopped"},
		//{"network", quadletSchemas["network"], "NetworkName", "foo", "podman network create foo"},
		{"network", quadletSchemas["network"], "Options", "isolate=true", "--opt isolate=true"},
		{"network", quadletSchemas["network"], "PodmanArgs", "--dns=192.168.55.1", "--dns=192.168.55.1"},
		{"network", quadletSchemas["network"], "Subnet", "192.5.0.0/16", "--subnet 192.5.0.0/16"},
	}
	// The execution loop
	for _, tt := range tests {
		quadletOpt := fmt.Sprintf("%s=%s", tt.key, tt.value)
		t.Run(quadletOpt, func(t *testing.T) {
			ans, err := QuadletOptionToPodman(tt.qType, tt.options, tt.key, tt.value)
			if err != nil {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestPodQuadletOptionsToPodmanTableDriven(t *testing.T) {
	quadletSchemas := map[string]map[string]SchemaOption{}
	quadletSchemas["volume"] = GetQuadletOptionsMap("volume")
	quadletSchemas["network"] = GetQuadletOptionsMap("network")
	quadletSchemas["container"] = GetQuadletOptionsMap("container")
	quadletSchemas["pod"] = GetQuadletOptionsMap("pod")

	// Defining the columns of the table
	var tests = []struct {
		qType   string
		options map[string]SchemaOption
		key     string
		value   string
		want    string
	}{
		// the table itself
		{"pod", quadletSchemas["pod"], "AddHost", "example.com:192.168.10.11", "--add-host example.com:192.168.10.11"},
		{"pod", quadletSchemas["pod"], "ContainersConfModule", "/etc/nvd.conf", "--module=/etc/nvd.conf"},
		{"pod", quadletSchemas["pod"], "DNS", "192.168.55.1", "--dns=192.168.55.1"},
		{"pod", quadletSchemas["pod"], "DNSOption", "ndots:1", "--dns-option=ndots:1"},
		{"pod", quadletSchemas["pod"], "DNSSearch", "example.com", "--dns-search example.com"},
		{"pod", quadletSchemas["pod"], "ExitPolicy", "stop", "--exit-policy stop"},
		{"pod", quadletSchemas["pod"], "GIDMap", "0:10000:10", "--gidmap=0:10000:10"},
		{"pod", quadletSchemas["pod"], "GlobalArgs", "--log-level=debug", "--log-level=debug"},
		{"pod", quadletSchemas["pod"], "HostName", "name", "--hostname=name"},
		{"pod", quadletSchemas["pod"], "IP", "192.5.0.1", "--ip 192.5.0.1"},
		{"pod", quadletSchemas["pod"], "IP6", "2001:db8::1", "--ip6 2001:db8::1"},
		{"pod", quadletSchemas["pod"], "Label", "\"XYZ\"", "--label \"XYZ\""},
		{"pod", quadletSchemas["pod"], "Network", "host", "--network host"},
		{"pod", quadletSchemas["pod"], "NetworkAlias", "name", "--network-alias name"},
		{"pod", quadletSchemas["pod"], "PodmanArgs", "--cpus=2", "--cpus=2"},
		{"pod", quadletSchemas["pod"], "PodName", "name", "--name=name"},
		{"pod", quadletSchemas["pod"], "PublishPort", "8080:80", "--publish 8080:80"},
		//{"pod", quadletSchemas["pod"], "ServiceName", "name", "Name the systemd unit name.service"},
		{"pod", quadletSchemas["pod"], "ShmSize", "100m", "--shm-size=100m"},
		{"pod", quadletSchemas["pod"], "SubGIDMap", "gtest", "--subgidname=gtest"},
		{"pod", quadletSchemas["pod"], "SubUIDMap", "utest", "--subuidname=utest"},
		{"pod", quadletSchemas["pod"], "UIDMap", "0:10000:10", "--uidmap=0:10000:10"},
		{"pod", quadletSchemas["pod"], "UserNS", "keep-id:uid=200,gid=210", "--userns keep-id:uid=200,gid=210"},
		{"pod", quadletSchemas["pod"], "Volume", "/source:/dest", "--volume /source:/dest"},
	}
	// The execution loop
	for _, tt := range tests {
		quadletOpt := fmt.Sprintf("%s=%s", tt.key, tt.value)
		t.Run(quadletOpt, func(t *testing.T) {
			ans, err := QuadletOptionToPodman(tt.qType, tt.options, tt.key, tt.value)
			if err != nil {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestContainerQuadletOptionsToPodmanTableDriven(t *testing.T) {
	quadletSchemas := map[string]map[string]SchemaOption{}
	quadletSchemas["volume"] = GetQuadletOptionsMap("volume")
	quadletSchemas["network"] = GetQuadletOptionsMap("network")
	quadletSchemas["container"] = GetQuadletOptionsMap("container")
	quadletSchemas["pod"] = GetQuadletOptionsMap("pod")

	// Defining the columns of the table
	var tests = []struct {
		qType   string
		options map[string]SchemaOption
		key     string
		value   string
		want    string
	}{
		// the table itself
		{"container", quadletSchemas["container"], "AddCapability", "CAP_SYS_ADMIN", "--cap-add CAP_SYS_ADMIN"},
		{"container", quadletSchemas["container"], "AddDevice", "/dev/net/tun", "--device /dev/net/tun"},
		{"container", quadletSchemas["container"], "AddHost", "example.com:192.168.10.11", "--add-host example.com:192.168.10.11"},
		{"container", quadletSchemas["container"], "Annotation", "XYZ", "--annotation XYZ"},
		{"container", quadletSchemas["container"], "AppArmor", "alternate-profile", "--security-opt apparmor=alternate-profile"},
		{"container", quadletSchemas["container"], "AutoUpdate", "registry", "--label io.containers.autoupdate=registry"},
		{"container", quadletSchemas["container"], "CgroupsMode", "no-conmon", "--cgroups no-conmon"},
		{"container", quadletSchemas["container"], "ContainersConfModule", "/etc/nvd.conf", "--module /etc/nvd.conf"},
		{"container", quadletSchemas["container"], "ContainerName", "name", "--name name"},
		{"container", quadletSchemas["container"], "DNS", "192.168.55.1", "--dns 192.168.55.1"},
		{"container", quadletSchemas["container"], "DNSOption", "ndots:1", "--dns-option ndots:1"},
		{"container", quadletSchemas["container"], "DNSSearch", "example.com", "--dns-search example.com"},
		{"container", quadletSchemas["container"], "DropCapability", "CAP_SYS_ADMIN", "--cap-drop CAP_SYS_ADMIN"},
		{"container", quadletSchemas["container"], "Entrypoint", "\"/bin/sh\"", "--entrypoint \"/bin/sh\""},
		{"container", quadletSchemas["container"], "Environment", "XYZ", "--env XYZ"},
		{"container", quadletSchemas["container"], "EnvironmentFile", "/etc/env", "--env-file /etc/env"},
		{"container", quadletSchemas["container"], "EnvironmentHost", "true", "--env-host"},
		{"container", quadletSchemas["container"], "Exec", "/bin/sh", "/bin/sh"},
		{"container", quadletSchemas["container"], "ExposeHostPort", "8080", "--expose 8080"},
		{"container", quadletSchemas["container"], "GIDMap", "0:10000:10", "--gidmap 0:10000:10"},
		{"container", quadletSchemas["container"], "GlobalArgs", "--log-level=debug", "--log-level=debug"},
		{"container", quadletSchemas["container"], "Group", "192", "--user :192"},
		{"container", quadletSchemas["container"], "GroupAdd", "keep-groups", "--group-add keep-groups"},
		{"container", quadletSchemas["container"], "HealthCmd", "\"curl -f http://localhost/ || exit 1\"", "--health-cmd \"curl -f http://localhost/ || exit 1\""},
		{"container", quadletSchemas["container"], "HealthInterval", "5m", "--health-interval 5m"},

		{"container", quadletSchemas["container"], "HealthLogDestination", "/foo/log", "--health-log-destination /foo/log"},
		{"container", quadletSchemas["container"], "HealthMaxLogCount", "5", "--health-max-log-count 5"},
		{"container", quadletSchemas["container"], "HealthMaxLogSize", "500", "--health-max-log-size 500"},

		{"container", quadletSchemas["container"], "HealthOnFailure", "kill", "--health-on-failure kill"},
		{"container", quadletSchemas["container"], "HealthRetries", "5", "--health-retries 5"},
		{"container", quadletSchemas["container"], "HealthStartPeriod", "1m", "--health-start-period 1m"},

		{"container", quadletSchemas["container"], "HealthStartupCmd", "command", "--health-startup-cmd command"},
		{"container", quadletSchemas["container"], "HealthStartupInterval", "1m", "--health-startup-interval 1m"},
		{"container", quadletSchemas["container"], "HealthStartupRetries", "8", "--health-startup-retries 8"},
		{"container", quadletSchemas["container"], "HealthStartupSuccess", "2", "--health-startup-success 2"},
		{"container", quadletSchemas["container"], "HealthStartupTimeout", "1m33s", "--health-startup-timeout 1m33s"},

		{"container", quadletSchemas["container"], "HealthTimeout", "2m", "--health-timeout 2m"},
		{"container", quadletSchemas["container"], "HostName", "name", "--hostname name"},
		{"container", quadletSchemas["container"], "HttpProxy", "true", "--http-proxy"},
		{"container", quadletSchemas["container"], "Image", "quay.io/centos/centos:latest", "quay.io/centos/centos:latest"},
		{"container", quadletSchemas["container"], "IP", "192.5.0.1", "--ip 192.5.0.1"},
		{"container", quadletSchemas["container"], "IP6", "2001:db8::1", "--ip6 2001:db8::1"},
		{"container", quadletSchemas["container"], "Label", "\"XYZ\"", "--label \"XYZ\""},
		{"container", quadletSchemas["container"], "LogDriver", "journald", "--log-driver journald"},
		{"container", quadletSchemas["container"], "LogOpt", "path=/var/log/container.log", "--log-opt path=/var/log/container.log"},
		{"container", quadletSchemas["container"], "Mask", "/sys/firmware", "--security-opt mask=/sys/firmware"},

		{"container", quadletSchemas["container"], "Memory", "20g", "--memory 20g"},

		{"container", quadletSchemas["container"], "Mount", "type=bind,source=/var/lib/db,destination=/var/lib/db", "--mount type=bind,source=/var/lib/db,destination=/var/lib/db"},
		{"container", quadletSchemas["container"], "Network", "host", "--network host"},
		{"container", quadletSchemas["container"], "NetworkAlias", "name", "--network-alias name"},
		{"container", quadletSchemas["container"], "NoNewPrivileges", "true", "--security-opt no-new-privileges"},
		{"container", quadletSchemas["container"], "Notify", "healthy", "--sdnotify healthy"},
		{"container", quadletSchemas["container"], "PidsLimit", "10", "--pids-limit 10"},
		{"container", quadletSchemas["container"], "Pod", "podname.service", "--pod podname.service"},
		{"container", quadletSchemas["container"], "PodmanArgs", "--cpus=2", "--cpus=2"},
		{"container", quadletSchemas["container"], "PublishPort", "8080:80", "--publish 8080:80"},
		{"container", quadletSchemas["container"], "Pull", "always", "--pull always"},
		{"container", quadletSchemas["container"], "ReadOnly", "true", "--read-only"},
		{"container", quadletSchemas["container"], "ReadOnlyTmpfs", "true", "--read-only-tmpfs"},

		{"container", quadletSchemas["container"], "ReloadCmd", "/usr/bin/command", ""},
		{"container", quadletSchemas["container"], "ReloadSignal", "SIGHUP", ""},
		{"container", quadletSchemas["container"], "Retry", "5", "--retry 5"},
		{"container", quadletSchemas["container"], "RetryDelay", "5s", "--retry-delay 5s"},

		{"container", quadletSchemas["container"], "Rootfs", "/var/lib/rootfs", "--rootfs /var/lib/rootfs"},
		{"container", quadletSchemas["container"], "RunInit", "true", "--init"},
		{"container", quadletSchemas["container"], "SeccompProfile", "/path/to/profile.json", "--security-opt seccomp=/path/to/profile.json"},
		{"container", quadletSchemas["container"], "Secret", "secret[,opt=opt]", "--secret secret[,opt=opt]"},
		{"container", quadletSchemas["container"], "SecurityLabelDisable", "true", "--security-opt label=disable"},
		{"container", quadletSchemas["container"], "SecurityLabelFileType", "user_home_t", "--security-opt label=filetype:user_home_t"},
		{"container", quadletSchemas["container"], "SecurityLabelLevel", "s0:c1,c2", "--security-opt label=level:s0:c1,c2"},
		{"container", quadletSchemas["container"], "SecurityLabelNested", "true", "--security-opt label=nested"},
		{"container", quadletSchemas["container"], "SecurityLabelType", "spc_t", "--security-opt label=type:spc_t"},
		{"container", quadletSchemas["container"], "ServiceName", "name", ""},
		{"container", quadletSchemas["container"], "ShmSize", "100m", "--shm-size 100m"},

		{"container", quadletSchemas["container"], "StartWithPod", "true", ""},
		{"container", quadletSchemas["container"], "StopSignal", "SIGINT", "--stop-signal SIGINT"},

		{"container", quadletSchemas["container"], "StopTimeout", "10", "--stop-timeout 10"},
		{"container", quadletSchemas["container"], "SubGIDMap", "gtest", "--subgidname gtest"},
		{"container", quadletSchemas["container"], "SubUIDMap", "utest", "--subuidname utest"},
		{"container", quadletSchemas["container"], "Sysctl", "net.ipv4.ip_forward=1", "--sysctl net.ipv4.ip_forward=1"},
		{"container", quadletSchemas["container"], "Timezone", "local", "--tz local"},
		{"container", quadletSchemas["container"], "Tmpfs", "/work", "--tmpfs /work"},
		{"container", quadletSchemas["container"], "UIDMap", "0:10000:10", "--uidmap 0:10000:10"},
		{"container", quadletSchemas["container"], "Ulimit", "nofile=100:200", "--ulimit nofile=100:200"},
		{"container", quadletSchemas["container"], "Unmask", "ALL", "--security-opt unmask=ALL"},
		{"container", quadletSchemas["container"], "User", "123", "--user 123"},
		{"container", quadletSchemas["container"], "UserNS", "keep-id:uid=200,gid=210", "--userns keep-id:uid=200,gid=210"},
		{"container", quadletSchemas["container"], "Volume", "/source:/dest", "--volume /source:/dest"},
		{"container", quadletSchemas["container"], "WorkingDir", "/work", "--workdir /work"},
	}
	// The execution loop
	for _, tt := range tests {
		quadletOpt := fmt.Sprintf("%s=%s", tt.key, tt.value)
		t.Run(quadletOpt, func(t *testing.T) {
			ans, err := QuadletOptionToPodman(tt.qType, tt.options, tt.key, tt.value)
			if err != nil {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

/*

//The following test was used to generate an actual podman create command to validate manually that the generated command options
//  were accepted by the podman runtime and appeared in the podman inspect output as expected.
//  It is not meant to be run as part of the automated tests, but it can be used as a reference for future tests.

func TestContainerQuadletOptionsToInspectTableDriven(t *testing.T) {

	// Defining the columns of the table
	var tests = []struct {
		qType   string
		options map[string]SchemaOption
		key     string
		value   string
		want    string
	}{
		// the table itself
		{"container", quadletSchemas["container"], "AddCapability", "CAP_SYS_ADMIN", "--cap-add CAP_SYS_ADMIN"},
		{"container", quadletSchemas["container"], "AddDevice", "/dev/net/tun", "--device /dev/net/tun"},
		{"container", quadletSchemas["container"], "AddHost", "example.com:192.168.10.11", "--add-host example.com:192.168.10.11"},
		{"container", quadletSchemas["container"], "Annotation", "XYZ", "--annotation XYZ"},
		{"container", quadletSchemas["container"], "AppArmor", "alternate-profile", "--security-opt apparmor=alternate-profile"},
		{"container", quadletSchemas["container"], "AutoUpdate", "registry", "--label io.containers.autoupdate=registry"},
		{"container", quadletSchemas["container"], "CgroupsMode", "no-conmon", "--cgroups no-conmon"},
		{"container", quadletSchemas["container"], "ContainersConfModule", "/etc/nvd.conf", "--module /etc/nvd.conf"},
		{"container", quadletSchemas["container"], "ContainerName", "name", "--name name"},
		{"container", quadletSchemas["container"], "DNS", "192.168.55.1", "--dns 192.168.55.1"},
		{"container", quadletSchemas["container"], "DNSOption", "ndots:1", "--dns-option ndots:1"},
		{"container", quadletSchemas["container"], "DNSSearch", "example.com", "--dns-search example.com"},
		{"container", quadletSchemas["container"], "DropCapability", "CAP_SYS_ADMIN", "--cap-drop CAP_SYS_ADMIN"},
		{"container", quadletSchemas["container"], "Entrypoint", "\"/bin/sh\"", "--entrypoint \"/bin/sh\""},
		{"container", quadletSchemas["container"], "Environment", "XYZ", "--env XYZ"},
		{"container", quadletSchemas["container"], "EnvironmentFile", "/etc/env", "--env-file /etc/env"},
		{"container", quadletSchemas["container"], "EnvironmentHost", "true", "--env-host"},
		{"container", quadletSchemas["container"], "Exec", "/bin/sh", "/bin/sh"},
		{"container", quadletSchemas["container"], "ExposeHostPort", "8080", "--expose 8080"},
		{"container", quadletSchemas["container"], "GIDMap", "0:10000:10", "--gidmap 0:10000:10"},
		{"container", quadletSchemas["container"], "GlobalArgs", "--log-level=debug", "--log-level=debug"},
		{"container", quadletSchemas["container"], "Group", "192", "--user :192"},
		{"container", quadletSchemas["container"], "GroupAdd", "keep-groups", "--group-add keep-groups"},
		{"container", quadletSchemas["container"], "HealthCmd", "\"curl -f http://localhost/ || exit 1\"", "--health-cmd \"curl -f http://localhost/ || exit 1\""},
		{"container", quadletSchemas["container"], "HealthInterval", "5m", "--health-interval 5m"},

		{"container", quadletSchemas["container"], "HealthLogDestination", "/foo/log", "--health-log-destination /foo/log"},
		{"container", quadletSchemas["container"], "HealthMaxLogCount", "5", "--health-max-log-count 5"},
		{"container", quadletSchemas["container"], "HealthMaxLogSize", "500", "--health-max-log-size 500"},

		{"container", quadletSchemas["container"], "HealthOnFailure", "kill", "--health-on-failure kill"},
		{"container", quadletSchemas["container"], "HealthRetries", "5", "--health-retries 5"},
		{"container", quadletSchemas["container"], "HealthStartPeriod", "1m", "--health-start-period 1m"},

		{"container", quadletSchemas["container"], "HealthStartupCmd", "command", "--health-startup-cmd command"},
		{"container", quadletSchemas["container"], "HealthStartupInterval", "1m", "--health-startup-interval 1m"},
		{"container", quadletSchemas["container"], "HealthStartupRetries", "8", "--health-startup-retries 8"},
		{"container", quadletSchemas["container"], "HealthStartupSuccess", "2", "--health-startup-success 2"},
		{"container", quadletSchemas["container"], "HealthStartupTimeout", "1m33s", "--health-startup-timeout 1m33s"},

		{"container", quadletSchemas["container"], "HealthTimeout", "2m", "--health-timeout 2m"},
		{"container", quadletSchemas["container"], "HostName", "name", "--hostname name"},
		{"container", quadletSchemas["container"], "HttpProxy", "true", "--http-proxy"},
		{"container", quadletSchemas["container"], "Image", "quay.io/centos/centos:latest", "quay.io/centos/centos:latest"},
		{"container", quadletSchemas["container"], "IP", "192.5.0.1", "--ip 192.5.0.1"},
		{"container", quadletSchemas["container"], "IP6", "2001:db8::1", "--ip6 2001:db8::1"},
		{"container", quadletSchemas["container"], "Label", "\"XYZ\"", "--label \"XYZ\""},
		{"container", quadletSchemas["container"], "LogDriver", "journald", "--log-driver journald"},
		{"container", quadletSchemas["container"], "LogOpt", "path=/var/log/container.log", "--log-opt path=/var/log/container.log"},
		{"container", quadletSchemas["container"], "Mask", "/sys/firmware", "--security-opt mask=/sys/firmware"},

		{"container", quadletSchemas["container"], "Memory", "20g", "--memory 20g"},

		{"container", quadletSchemas["container"], "Mount", "type=bind,source=/var/lib/db,destination=/var/lib/db", "--mount type=bind,source=/var/lib/db,destination=/var/lib/db"},
		{"container", quadletSchemas["container"], "Network", "host", "--network host"},
		{"container", quadletSchemas["container"], "NetworkAlias", "name", "--network-alias name"},
		{"container", quadletSchemas["container"], "NoNewPrivileges", "true", "--security-opt no-new-privileges"},
		{"container", quadletSchemas["container"], "Notify", "healthy", "--sdnotify healthy"},
		{"container", quadletSchemas["container"], "PidsLimit", "10", "--pids-limit 10"},
		{"container", quadletSchemas["container"], "Pod", "podname.service", "--pod podname.service"},
		{"container", quadletSchemas["container"], "PodmanArgs", "--cpus=2", "--cpus=2"},
		{"container", quadletSchemas["container"], "PublishPort", "8080:80", "--publish 8080:80"},
		{"container", quadletSchemas["container"], "Pull", "always", "--pull always"},
		{"container", quadletSchemas["container"], "ReadOnly", "true", "--read-only"},
		{"container", quadletSchemas["container"], "ReadOnlyTmpfs", "true", "--read-only-tmpfs"},

		{"container", quadletSchemas["container"], "ReloadCmd", "/usr/bin/command", ""},
		{"container", quadletSchemas["container"], "ReloadSignal", "SIGHUP", ""},
		{"container", quadletSchemas["container"], "Retry", "5", "--retry 5"},
		{"container", quadletSchemas["container"], "RetryDelay", "5s", "--retry-delay 5s"},

		{"container", quadletSchemas["container"], "Rootfs", "/var/lib/rootfs", "--rootfs /var/lib/rootfs"},
		{"container", quadletSchemas["container"], "RunInit", "true", "--init"},
		{"container", quadletSchemas["container"], "SeccompProfile", "/path/to/profile.json", "--security-opt seccomp=/path/to/profile.json"},
		{"container", quadletSchemas["container"], "Secret", "secret,opt=opt", "--secret secret,opt=opt"},
		{"container", quadletSchemas["container"], "SecurityLabelDisable", "true", "--security-opt label=disable"},
		{"container", quadletSchemas["container"], "SecurityLabelFileType", "user_home_t", "--security-opt label=filetype:user_home_t"},
		{"container", quadletSchemas["container"], "SecurityLabelLevel", "s0:c1,c2", "--security-opt label=level:s0:c1,c2"},
		{"container", quadletSchemas["container"], "SecurityLabelNested", "true", "--security-opt label=nested"},
		{"container", quadletSchemas["container"], "SecurityLabelType", "spc_t", "--security-opt label=type:spc_t"},
		{"container", quadletSchemas["container"], "ServiceName", "name", ""},
		{"container", quadletSchemas["container"], "ShmSize", "100m", "--shm-size 100m"},

		{"container", quadletSchemas["container"], "StartWithPod", "true", ""},
		{"container", quadletSchemas["container"], "StopSignal", "SIGINT", "--stop-signal SIGINT"},

		{"container", quadletSchemas["container"], "StopTimeout", "10", "--stop-timeout 10"},
		{"container", quadletSchemas["container"], "SubGIDMap", "gtest", "--subgidname gtest"},
		{"container", quadletSchemas["container"], "SubUIDMap", "utest", "--subuidname utest"},
		{"container", quadletSchemas["container"], "Sysctl", "net.ipv4.ip_forward=1", "--sysctl net.ipv4.ip_forward=1"},
		{"container", quadletSchemas["container"], "Timezone", "local", "--tz local"},
		{"container", quadletSchemas["container"], "Tmpfs", "/work", "--tmpfs /work"},
		{"container", quadletSchemas["container"], "UIDMap", "0:10000:10", "--uidmap 0:10000:10"},
		{"container", quadletSchemas["container"], "Ulimit", "nofile=100:200", "--ulimit nofile=100:200"},
		{"container", quadletSchemas["container"], "Unmask", "ALL", "--security-opt unmask=ALL"},
		{"container", quadletSchemas["container"], "User", "123", "--user 123"},
		{"container", quadletSchemas["container"], "UserNS", "keep-id:uid=200,gid=210", "--userns keep-id:uid=200,gid=210"},
		{"container", quadletSchemas["container"], "Volume", "/source:/dest", "--volume /source:/dest"},
		{"container", quadletSchemas["container"], "WorkingDir", "/work", "--workdir /work"},
	}
	// The execution loop
	cmd := []string{"podman", "container", "create"}
	for _, tt := range tests {
		quadletOpt := fmt.Sprintf("%s=%s", tt.key, tt.value)
		t.Run(quadletOpt, func(t *testing.T) {
			ans, err := quadletOptionToPodman(tt.qType, tt.options, tt.key, tt.value)
			if err != nil {
				t.Errorf("got %s, want %s", ans, tt.want)
			} else {
				cmd = append(cmd, ans)
			}
		})
	}
	cmd = append(cmd, "docker.io/library/ubuntu")
	cmd = append(cmd, "testoptions")

	fmt.Println(strings.Join(cmd, " "))
}
*/
