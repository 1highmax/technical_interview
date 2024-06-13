# Exercise 1

## Task Description
Bootable Linux image via QEMU

In this exercise you are expected to create a shell script that will run in a Linux environment (will be tested on Ubuntu 20.04 LTS or 22.04 LTS). This shell script should create and run an AMD64 Linux filesystem image using QEMU that will print “hello world” after successful startup. Bonus points for creating a fully bootable filesystem image (but not mandatory). The system shouldn’t contain any user/session management or prompt for login information to access the filesystem.  

You can use any version/flavor of the Linux kernel. The script can either download and build the kernel from source on the host environment or download a publicly available pre-built kernel.

The script shouldn’t ask for any user input unless superuser privileges are necessary for some functionality, therefore any additional information that you require for the script should be available in your repository.
The script should run within the working directory and not consume any other locations on the host file system.


## Assumptions
- The instructions do not state the host architecture. To avoid installing cross-compilation toolchains and checking host architecture, I use a prebuilt x86_64 kernel, and a prebuilt busybox amd64 [binary](busybox) (It's also more convencient for me working on my aarch64 machine)
- I assume all dependencies to be apt-based, since nothing regarding snap packages or installations from source are mentioned
- I assume the sudo command to be available
- Since performance does not matter in this project, and to avoid additional dependency checks, I do not use KVM

## Usage
```bash
./runQemuHelloWorld.bash
```

## Discussion
- I use busybox sh is an init process, and cttyhack to avoid the warning "job control turned off"
- I include busybox in the repo, since its only 1M and the busybox download server is unreliable
- This has been tested in fresh multipass ubuntu 22.04