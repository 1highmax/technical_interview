#!/bin/bash

# Do not pop up menuconfigs telling about available kernel updates
export DEBIAN_FRONTEND=noninteractive

# Define the QEMU command, add action poweroff to allow graceful shutdown
export QEMU_COMMAND="qemu-system-x86_64 -kernel kernel/boot/vmlinuz-6.9.0-060900-generic -hda disk.img -nographic -serial mon:stdio -append 'root=/dev/sda console=ttyS0 init=/init' -no-reboot -no-shutdown -machine pc -action shutdown=poweroff"

# Clean up any previous files
rm -rf kernel initramfs* linux* CHECKSUMS* ubuntu-key.asc rootfs disk.img &> /dev/null 

# Enable debugging if the script is run with -v
if [[ "$1" == "-v" ]]; then
    set -x
fi

# Function to print in green color
print_green() {
    echo -e "\e[32m$1\e[0m"
}

# Function to print in red color
print_red() {
    echo -e "\e[31m$1\e[0m"
}

# Checking if the user is root to not have sudo errors
if [ "$(id -u)" -eq 0 ]; then
    print_red "Do not run this script as root"
    exit 1
else
    print_green "Executing as non-root user"
fi

# Install required packages
sudo apt update
sudo apt install -y wget gnupg qemu-system cpio

# Define URLs and files
KERNEL_URL="https://kernel.ubuntu.com/~kernel-ppa/mainline/v6.9/amd64/linux-image-unsigned-6.9.0-060900-generic_6.9.0-060900.202405122134_amd64.deb"
CHECKSUMS_URL="https://kernel.ubuntu.com/~kernel-ppa/mainline/v6.9/amd64/CHECKSUMS"
CHECKSUMS_GPG_URL="https://kernel.ubuntu.com/~kernel-ppa/mainline/v6.9/amd64/CHECKSUMS.gpg"
KERNEL_FILE="linux-image-unsigned-6.9.0-060900-generic_6.9.0-060900.202405122134_amd64.deb"

# Download files
wget -O $KERNEL_FILE $KERNEL_URL
wget -O CHECKSUMS $CHECKSUMS_URL
wget -O CHECKSUMS.gpg $CHECKSUMS_GPG_URL

# Import the Ubuntu Kernel PPA key
gpg --keyserver hkps://keyserver.ubuntu.com --recv-keys 60AA7B6F30434AE68E569963E50C6A0917C622B0

# Verify the GPG signature
gpg --verify CHECKSUMS.gpg CHECKSUMS
if [ $? -ne 0 ]; then
    print_red "GPG signature verification failed"
    exit 1
else
    print_green "GPG signature verification succeeded"
fi

# Verify the SHA256 hash
echo "Checking the SHA256 hash..."
grep $(basename $KERNEL_FILE) CHECKSUMS | shasum -c -
if [ $? -ne 0 ]; then
    print_red "Downloaded kernel file has an invalid hash"
    exit 1
else
    print_green "Downloaded kernel file has a valid hash"
fi

# Extract the kernel image
mkdir -p kernel
dpkg-deb -x $KERNEL_FILE kernel

# Create root filesystem with a simple init script
mkdir -p rootfs/{bin,sbin,etc,proc,sys,usr/bin,usr/sbin,dev,lib}

# Copy the provided BusyBox binary to rootfs
cp busybox rootfs/bin/

# Create the symlink to sh
( cd rootfs/bin && ln -s busybox sh )
( cd rootfs/bin && ln -s busybox mount )
( cd rootfs/bin && ln -s busybox grep )
( cd rootfs/bin && ln -s busybox ln )
( cd rootfs/bin && ln -s busybox poweroff )


# Write the QEMU command to a file in the root filesystem
echo $QEMU_COMMAND > rootfs/qemu_command.txt

cat > rootfs/init << 'EOF'
#!/bin/sh
handle_shutdown() {
    echo "Shutting down..."
    sync
    poweroff -f
}

trap 'handle_shutdown' SIGTERM
trap 'handle_shutdown' EXIT  # Trap shell exit to handle Ctrl+D

mount -t proc none /proc
mount -t sysfs none /sys
mount -o remount,rw /

# Symlink busybox applets
for applet in $(/bin/busybox --list); do
    ( cd /bin && /bin/busybox ln -s /bin/busybox $applet &>/dev/null)
done

# Switch to an interactive shell
echo "####################################################################################"
echo "Hello, World!"
echo "This is busybox shell, running in a minimal root filesystem. Ctrl-D to shutdown."
echo "To run this fully bootable VM disk again after shutdown, see /qemu_command.txt:"
cat /qemu_command.txt
echo "####################################################################################"
setsid cttyhack sh
EOF

chmod +x rootfs/init

# Ensure all files have correct permissions
chmod -R 755 rootfs

# Verify the init script
if [ -f rootfs/init ] && [ -x rootfs/init ]; then
    print_green "Init script is in place and executable."
else
    print_red "Init script is missing or not executable."
    exit 1
fi

# Create a directory for mounting the disk image
sudo mkdir -p /mnt/disk

# Create a disk image and format it
dd if=/dev/zero of=disk.img bs=1M count=64
mkfs.ext4 -F disk.img

# Mount the disk image and copy the root filesystem
sudo mount -o loop disk.img /mnt/disk
sudo cp -r rootfs/* /mnt/disk
sudo umount /mnt/disk

# Check the file structure inside the disk image
sudo mount -o loop disk.img /mnt/disk
if [ -f /mnt/disk/init ] && [ -x /mnt/disk/init ]; then
    print_green "Init script is correctly placed in the disk image."
else
    print_red "Init script is not correctly placed in the disk image."
    sudo umount /mnt/disk
    exit 1
fi
sudo umount /mnt/disk

# Run QEMU 
sh -c "$QEMU_COMMAND"

# After QEMU exits, ensure the disk image is unmounted
if mountpoint -q /mnt/disk; then
    sudo umount /mnt/disk
fi
