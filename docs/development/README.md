# Development environment

## 1. Install dependencies and Go runtime

### For Debian/Ubuntu
```bash
apt-get install -y mercurial git-core wget make
wget https://storage.googleapis.com/golang/go1.7.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go*-*.tar.gz
```

### For OSX using binary package
```bash
wget https://storage.googleapis.com/golang/go1.7.1.darwin-amd64.tar.gz
sudo tar -C /usr/local -xzf go*-*.tar.gz
```

### For OSX if you have brew.sh
```
brew install go
```

### For OSX using installation package
```
wget https://storage.googleapis.com/golang/go1.7.1.darwin-amd64.pkg
open go*-*.pkg
```

### For FreeBSD
```
pkg install go-1.7.1 gmake git mercurial
```

## 2. Install Docker Engine

The Docker Engine is required to create pre-built image that is embedded into runner and loaded when using docker executor.

Make sure that on machine that is running your Docker Engine you have a `binfmt_misc`.
This is required to be able to build ARM images that are embedded into GitLab Runner binary.
 
* For Debian/Ubuntu it's sufficient to execute:
    
    ```
    apt-get install binfmt-support qemu-user-static
    ```

* For Docker for MacOS/Windows `binfmt_misc` is enabled by default.

* For CoreOS (but also works on Debian and Ubuntu) you need to execute the following script on system start:

    ```
    #!/bin/sh
    
    set -xe
    
    /sbin/modprobe binfmt_misc
    
    mount -t binfmt_misc binfmt_misc /proc/sys/fs/binfmt_misc
    
    # Support for ARM binaries through Qemu:
    { echo ':arm:M::\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:/usr/bin/qemu-arm-static:' > /proc/sys/fs/binfmt_misc/register; } 2>/dev/null
    { echo ':armeb:M::\x7fELF\x01\x02\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff:/usr/bin/qemu-armeb-static:' > /proc/sys/fs/binfmt_misc/register; } 2>/dev/null
    ```

[Install Docker Engine](https://docs.docker.com/engine/installation/)

## 3. Configure Go

Add to `.profile` or `.bash_profile`:

```bash
export GOPATH=$HOME/Go
export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin
```

Create new terminal session and create $GOPATH directory:

```
mkdir -p $GOPATH
```

## 4. Download runner sources

```
go get gitlab.com/gitlab-org/gitlab-ci-multi-runner
cd $GOPATH/src/gitlab.com/gitlab-org/gitlab-ci-multi-runner/
```

## 5. Install runner dependencies

This will download and restore all dependencies required to build runner:
```
make deps
```

**For FreeBSD use `gmake deps`**

## 6. Run runner

Normally you would use `gitlab-runner`, in order to compile and run Go source use go toolchain:

```
make install
gitlab-ci-multi-runner run
```

You can run runner in debug-mode:

```
make install
gitlab-ci-multi-runner --debug run
```

## 7. Compile and install runner binary as `gitlab-ci-multi-runner`

```
make install
```

## 8. Run test suite locally

GitLab Runner test suite consists of "core" tests and tests for executors.
Tests for executors require certain binaries to be installed on your local
machine. Some of these binaries cannot be installed on all operating
systems. If a binary is not installed tests requiring this binary will be
skipped.

These are the binaries that you can install:
1. [VirtualBox](https://www.virtualbox.org/wiki/Downloads) and [Vagrant](https://www.vagrantup.com/downloads.html)
1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/) with
  [Minikube](https://github.com/kubernetes/minikube)
1. [Parallels](http://www.parallels.com/products/desktop/download/)
1. [PowerShell](https://msdn.microsoft.com/en-us/powershell)

After installing the binaries run:

```
make development_setup
```

To execute the tests run:

```
make test
```

## 9. Congratulations!

You can start hacking GitLab-Runner code. If you are interested you can use Intellij IDEA Community Edition with [go-lang-idea-plugin](https://github.com/go-lang-plugin-org/go-lang-idea-plugin) to edit and debug code.

## Managing build dependencies

GitLab Runner uses [Govendor](https://github.com/kardianos/govendor) to manage
its dependencies - they get checked into the repository under the `vendor/` directory,
with a manifest stored in `vendor/vendor.json`.

If your contribution adds, removes or updates any dependencies to the runner,
please ensure the vendored copies updated in lock-step.

**For added/removed dependencies:**

1. Run `go get [package]` (if you haven't done this already) to download required package into your `$GOPATH`.
   Repeat this for any added package.

1. Commit all added changes and stash/remove all uncommited changes to have a clean working directory.

1. Execute

    ```bash
    $ make update_govendor_dependencies
    ```

    This task will:
    * remove all unused packages with,
    * add all missing or external packages (packages that are not stored in vendor but can be found in `$GOPATH`),
    * show git status and ask if changes in `vendor/` or done properly,
    * show diff of `vendor/vendor.json` and ask if it is done properly,
    * add and commit changes in `vendor/` directory.

**For dependencies that need update:**

1. Run `go get -u [package]` (if you haven't done this already) to upload selected package present in your `$GOPATH`
   or download it if it's not present.

    If you want to use specific version then go to the package directory in your `$GOPATH` and checkout it to this
    version, e.g.:

    ```bash
    $ cd $GOPATH/src/github.com/docker/docker/client
    $ git checkout v1.13.0
    ```

    Repeat this for any added package.

1. Execute

    ```bash
    $ make update_govendor_dependencies
    ```

    This task will behave just like described above.

## Troubleshooting

### executor_docker.go missing Asset symbol

This error happens due to missing executors/docker/bindata.go file that is generated from docker prebuilts.
Which is especially tricky on Windows.

Try to execute: `make deps docker`, if it doesn't help you can do that in steps:
1. Execute `go get -u github.com/jteeuwen/go-bindata/...`
2. Download https://gitlab-ci-multi-runner-downloads.s3.amazonaws.com/master/docker/prebuilt-x86_64.tar.xz and save to out/docker/prebuilt-x86_64.tar.xz
3. Download https://gitlab-ci-multi-runner-downloads.s3.amazonaws.com/master/docker/prebuilt-arm.tar.xz and save to out/docker/prebuilt-arm.tar.xz
4. Execute `make docker` or check the Makefile how this command looks like
