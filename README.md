# quadctl

A compose-like command-line tool to run Podman Quadlets locally without systemd and to facilitate installation and management of quadlets when using systemd is desired.

## Why?

Podman Quadlets use systemd to orchestrate and manage containers in the same way it does for all other services on modern Linux machines. That is an attractive proposition if you don't need the complex multi-server clustering features of, for example, kubernetes. You might also prefer the simple .ini file format over .yaml :)

However, if you're coming from docker compose, quadlets look complex. Multiple files need to be deployed to one or two (out of 10 possible) systemd quadlet generator directories and you have to get familiar with systemctl commands, daemon reloads, --user flag, journalctl, etc. To top it off, there is no way to run what you define in your quadlet files except for systemd, which causes many to treat quadlets as a late-stage 'production deployment' step if they don't give up entirely. Many have complained that they need a way to run "locally" for development before "deploying" to systemd.

Quadctl aims to provide a simple and consistent CLI for running and managing quadlets with and without systemd. 

## Demo

![Alt Text](./demo.gif)

## Features

* Unified command set for running directly or under systemd:
  * Use `quadctl start` to create and start _**rootless**_ containers directly under **podman**
  * Use `sudo quadctl start` to create and start _**rootful**_ containers directly under **podman**
  * Use `quadctl -s start` to create and start _**rootless**_ containers under **systemd**
  * Use `sudo quadctl -s start` to create and start _**rootful**_ containers under **systemd**
  * ... similarly for all other commands 
* Quadlet dependency ordering handled by quadctl when run directly, or by systemd when -s flag provided.
* Quadlet supports .container, .pod, .volume, .network and .quadlet (the currently proposed all-in-one .quadlet file format)
* Quadlet applications are organized in directories
  * e.g.
```
── /quadlet.src.path
   ├─ diun
   │  ╰─ diun.container
   ╰─ homebox
      ├─ homebox-app.container
      ├─ homebox-app.container.d
      │  ╰─ app.config
      ├─ homebox-data.volume
      ╰─ homebox.pod
```
*
  *  From /homebox, `quadctl start` works similarly to `docker compose up`
  *  From /quadlet.src.path, `quadctl start homebox` will bring up the app
  *  If quadlet.src.path is configured, `quadctl start homebox` will work from anywhere on the system
* Deploying to and removing from systemd quadlet generator directories is handled automatically when create and remove are used with the -s flag.
* Systemd reload is handled automatically
* The `list` command produces a tree listing of quadlets in quadlet.src.path or systemd quadlet generator directories.
* The `ps`,`stats`,`images`,`status` and `logs` commands are context-aware, providing results filtered to resources defined by the set of quadlets in the designated path. `status` and `logs` also invoke systemd status and journalctl when -s flag is provided.
* Supports the optional use of sub-directories in the systemd quadlet generator locations for better organization
* Supports the optional use of symbolic links in the systemd quadlet generator locations

## Installation

The below command line downloads the latest release and attempts to install to /usr/local/bin. Alternatively, go to the latest release page and manually download the tar file and extract to your preferred $PATH location for the binary. 

```
curl -sL github.com/fkmiec/quadctl/releases/latest/download/quadctl_linux_amd64.tar | sudo tar xv -C /usr/local/bin
```

On first invocation, quadctl will install a default quadctl.ini config file to ~/.config/quadctl. It is recommended that you review and update the location configurations to match your desired workflow: 

* quadlet.src.path - A directory location where subdirectories represent quadlet applications. Default is ~/.local/quadlets.
* quadlet.user.path - The systemd quadlet generator directory to use for rootless quadlets.
* quadlet.root.path - The systemd quadlet generator directory to use for rootful quadlets.

## Usage

```
Orchestrator for Podman Quadlets (with and without systemd)
Usage: quadctl [flags] <command> [path]

Flags:
  -s	Use systemd for managing services (default: false)
  -systemd
    	Use systemd for managing services (default: false)

Commands:
  pull       : Pull required images
  create     : Create resources (do not start). Use -s flag to generate quadlets under systemd.
  start      : Create (if missing) and start resources. Use -s flag to start under systemd.
  run        : Run a single .container in the foreground. Not supported for systemd. See quadctl run --help.
  stop       : Stop running services (do not remove). Use -s flag to stop under systemd.
  remove, rm : Remove stopped resources. Use -s flag to remove generated quadlets under systemd.
  logs       : Show logs of running containers. Use -s flag to view systemd logs.
  list, ls   : List quadlets in the configured quadlet_path or systemd path if -s flag is used.

Wrapper commands (filtered to defined resources):
  images : Show images defined for the set of related quadlets.
  ps     : Show state of containers.
  stats  : Show live stats for containers.
  status : Show current status. Use -s flag to see systemd status.

Requirements:
  A quadctl.ini config file is required. Default location is $HOME/.config/quadctl.
    A default config file will be created if not found.
  Set QUADCTL_CONFIG_DIR=<absolute path to config directory> in /etc/environment to
    change config location and/or ensure found when using sudo.

```

## A note on working with .quadlets files

Podman added support for a combined .quadlets file format in which you may combine the contents of invididual .container, .pod, .volume and .network quadlets. See the example on the [podman quadlet install documentation page](https://docs.podman.io/en/stable/markdown/podman-quadlet-install.1.html). This seems like a convenient and desireable option. However, the implementation appears limited to just the `podman quadlet install` command, which will extract the individual quadlets when installing them. The quadlet generator, and even the `podman quadlet rm` command, do not appear to recognize the .quadlets file extension, which can lead to confusion. 

Quadctl supports the .quadlets file format for all commands (ie. create, start, stop, rm, status, ps, etc.). It extracts the individual quadlets into their own files before processing and when the -s flag is provided, it _installs the extracted files_ to the quadlet generator directory. This is all essentially transparent, but you should be aware so that you are not confused when you invoke `quadctl -s ls` and see the extracted files, rather than your original .quadlets file. You'll also want to avoid using the `podman quadlet` subcommands and `quadctl` at the same time ... results in that case are likely to be frustrating. 

## A note on the use symbolic links configuration

Quadctl supports the optional use of symbolic links in the quadlet generator directories. Some quadlet users like this approach because they can edit their source quadlet files, reload systemd and restart to see the changes. The `quadctl -s start` command will reload systemd and restart the quadlets to support that workflow.

When the use_symbolic_links configuration is used in conjunction with use_subdirectories, the source quadlet directory is linked. `quadctl -s rm` will correctly remove the link without touching the original directory or its quadlet files. However, be careful not to manually perform a `rm -r` on the link as that will result in deletion of the source quadlet files. Likewise, using `podman quadlet rm` will delete the source quadlet file(s) in this case. For this reason, _*it's recommended that you avoid symbolic links.*_ With use_symbolic_links=false, a `quadctl -s rm` followed by `quadctl -s start` achieves a similar result with none of the risk.  
